// Copyleft (ɔ) 2018 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package cmd

import (
	"context"
	"errors"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	"github.com/CaliOpen/Caliopen/src/backend/main/go.backends/index/elasticsearch"
	"github.com/CaliOpen/Caliopen/src/backend/main/go.backends/store/cassandra"
	"github.com/CaliOpen/Caliopen/src/backend/main/go.main/facilities/REST"
	log "github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	"github.com/satori/go.uuid"
	"github.com/spf13/cobra"
	"gopkg.in/olivere/elastic.v5"
	"io"
	"time"
)

var identitiesMigrationCmd = &cobra.Command{
	Use:   "identitiesMigration",
	Short: "manage db and index migration for new model UserIdentity",
	Long: `command merges LocalIdentities and RemoteIdentities into new UserIdentities,
	builds lookup tables, then iterates over all messages in db and index to replace identities by user_identities`,
	Run: identitiesMigration,
}

func init() {
	RootCmd.AddCommand(identitiesMigrationCmd)
}

func identitiesMigration(cmd *cobra.Command, args []string) {
	// this script needs to read usernames from vault
	// replace apiv2's credentials with imapworker's credentials
	apiConf.BackendConfig.Settings.VaultSettings.Username = imapWorkerConf.StoreConfig.VaultConfig.Username
	apiConf.BackendConfig.Settings.VaultSettings.Password = imapWorkerConf.StoreConfig.VaultConfig.Password
	apiConf.BackendConfig.Settings.VaultSettings.Url = imapWorkerConf.StoreConfig.VaultConfig.Url

	var err error
	//check/get connexions to facilities
	//store
	var Store *store.CassandraBackend
	Store, err = getStoreFacility()
	if err != nil {
		log.WithError(err).Fatalf("initialization of %s backend failed", apiConf.BackendName)
	}
	defer Store.Close()
	//index
	var Index *index.ElasticSearchBackend
	Index, err = getIndexFacility()
	if err != nil {
		log.WithError(err).Fatalf("initialization of %s index failed", apiConf.IndexConfig.IndexName)
	}
	defer Index.Close()

	var Rest *REST.RESTfacility
	Rest, err = getRESTFacility()
	if err != nil {
		log.WithError(err).Fatal("initialization of ReST facility failed")
	}

	/* get all users id */
	erroneousUsers := []string{}
	erroneousMessages := []string{}
	if users, err := Store.Session.Query(`SELECT user_id, local_identities FROM user`).Iter().SliceMap(); err != nil {
		log.WithError(err).Fatal("failed to retrieve user ids")
	} else {
		users_total := len(users)
		users_count := 0
		log.Infof("\nFound %d users to work with\n", users_total)
		for _, user := range users {
			users_count++
			log.Infof("starting migration for user %d/%d", users_count, users_total)
			var userId UUID
			userId.UnmarshalBinary(user["user_id"].(gocql.UUID).Bytes())
			var localId *UUID
			var err error

			// check if a local user_identity has already been created
			localIds, err := Rest.RetrieveLocalIdentities(userId.String())
			if err != nil || len(localIds) == 0 {
				localId, err = createLocal(user, Store, erroneousUsers, userId, Rest)
				if err != nil {
					log.WithError(err).Warnf("createLocal returned error for user %s", userId.String())
				}
			} else {
				localId = &(localIds[0].Id)
			}

			err = createRemotes(Store, erroneousUsers, userId, Rest)
			if err != nil {
				log.WithError(err).Warnf("createRemote returned error for user %s", userId.String())
			}

			if localId != nil {
				err = updateSentMessages(Store, Index, userId, Rest, erroneousMessages, *localId)
				if err != nil {
					log.WithError(err).Warnf("updateSentMessages returned error for user %s", userId.String())
				}
			}
		}
		log.Info("\nAll done\n")
		if len(erroneousUsers) > 0 {
			log.Warnf("Users with errors : %v", erroneousUsers)
		}
		if len(erroneousMessages) > 0 {
			log.Warnf("Messages with errors : %v", erroneousMessages)
		}
	}
}

func createLocal(user map[string]interface{}, Store *store.CassandraBackend, erroneousUsers []string, userId UUID, Rest *REST.RESTfacility) (localId *UUID, err error) {
	log.Info("creating UserIdentity for local")
	local := map[string]interface{}{}
	identifier := user["local_identities"].([]string)
	if len(identifier) > 0 {
		err = Store.Session.Query(`SELECT * FROM local_identity WHERE identifier = ?`, identifier[0]).MapScan(local)
		if err != nil {
			log.WithError(err).Warnf("failed to retrieve local identity '%s' for user %s", identifier[0], userId.String())
			erroneousUsers = append(erroneousUsers, userId.String())
			return
		}
		uid := new(UUID)
		err = uid.UnmarshalBinary(uuid.NewV4().Bytes())
		if err != nil {
			log.WithError(err).Warnf("failed to create uuid for local identity for user %s", userId.String())
			erroneousUsers = append(erroneousUsers, userId.String())
			return
		}
		localUser := UserIdentity{
			DisplayName: local["display_name"].(string),
			Id:          *uid,
			Identifier:  local["identifier"].(string),
			Protocol:    "smtp",
			Status:      "active",
			Type:        "local",
			UserId:      userId,
		}
		caliopenErr := Rest.CreateUserIdentity(&localUser)
		if caliopenErr != nil {
			log.WithError(caliopenErr).Warnf("failed to save local identity %v for user %s", localUser, userId.String())
			erroneousUsers = append(erroneousUsers, userId.String())
			return nil, caliopenErr.Cause()
		}
		return uid, nil
	}
	return nil, errors.New("local_identities empty")
}

func createRemotes(Store *store.CassandraBackend, erroneousUsers []string, userId UUID, Rest *REST.RESTfacility) (err error) {
	log.Info("creating UserIdentity for remotes")
	remotes, err := Store.Session.Query(`SELECT * from remote_identity WHERE user_id = ?`, userId.String()).Iter().SliceMap()
	if err != nil {
		log.WithError(err).Warnf("failed to retrieve remote identities for user %s", userId.String())
		erroneousUsers = append(erroneousUsers, userId.String())
		return
	}

	for _, remote := range remotes {
		remoteID := new(UserIdentity)
		remoteID.NewEmpty()

		if dn, ok := remote["display_name"].(string); ok {
			remoteID.DisplayName = dn
		}

		if infos, ok := remote["infos"].(map[string]string); ok {
			remoteID.Infos = make(map[string]string)
			for k, v := range infos {
				remoteID.Infos[k] = v
			}
		}
		if lc, ok := remote["last_check"].(time.Time); ok {
			remoteID.LastCheck = lc
		}
		if remote_id, ok := remote["remote_id"].(gocql.UUID); ok {
			remoteID.Id.UnmarshalBinary(remote_id.Bytes())
		}
		if status, ok := remote["status"].(string); ok {
			remoteID.Status = status
		}
		if t, ok := remote["type"].(string); ok {
			remoteID.Protocol = t
		}
		if userid, ok := remote["user_id"].(gocql.UUID); ok {
			remoteID.UserId.UnmarshalBinary(userid.Bytes())
		}

		//try to fill-in identifier from username or display name
		if apiConf.BackendConfig.Settings.UseVault {
			cred, err := Store.RetrieveCredentials(userId.String(), remoteID.Id.String())
			if err != nil {
				remoteID.Identifier = remoteID.DisplayName
			} else {
				if username, ok := cred["username"]; ok {
					remoteID.Identifier = username
				} else {
					remoteID.Identifier = remoteID.DisplayName
				}
			}
		} else {
			if credentials, ok := remote["credentials"]; ok && credentials != nil {
				cred := &Credentials{}
				cred.UnmarshalCQLMap(credentials.(map[string]string))
				remoteID.Credentials = cred
				if username, ok := credentials.(map[string]string)["username"]; ok {
					remoteID.Identifier = username
				} else {
					remoteID.Identifier = remoteID.DisplayName
				}
			} else {
				remoteID.Identifier = remoteID.DisplayName
			}
		}

		remoteID.Type = RemoteIdentity
		caliopenErr := Rest.CreateUserIdentity(remoteID)
		if caliopenErr != nil {
			log.WithError(caliopenErr).Warnf("failed to save remote identity %v for user %s", *remoteID, userId.String())
			erroneousUsers = append(erroneousUsers, userId.String())
			continue
		}
	}
	return nil
}

func updateSentMessages(Store *store.CassandraBackend, Index *index.ElasticSearchBackend, userId UUID, Rest *REST.RESTfacility, erroneousMessages []string, localId UUID) error {
	log.Info("updating messages")
	/* get all sent messages id for user from index */
	scroll := Index.Client.Scroll(userId.String()).Type(MessageIndexType).
		Query(elastic.NewTermsQuery("is_received", false)).
		FetchSource(false).
		Size(100)
	for {
		results, err := scroll.Do(context.TODO())
		if err == io.EOF {
			return nil // all results retrieved for this user
		}
		if err != nil {
			// something went wrong
			log.WithError(err).Warnf("failed to scroll over indexed messages for user %s", userId.String())
			return err
		}
		log.Infof("%d messages to update", results.TotalHits())
		for _, hit := range results.Hits.Hits {
			message, err := Rest.GetMessage(userId.String(), hit.Id)
			if err != nil {
				continue
			}
			userIdentities := []UUID{localId}
			updatedFields := map[string]interface{}{"UserIdentities": userIdentities}
			message.UserIdentities = userIdentities
			err = Store.UpdateMessage(message, updatedFields)
			if err != nil {
				log.WithError(err).Warnf("failed to update message %s in db", hit.Id)
				erroneousMessages = append(erroneousMessages, userId.String())
			}
			err = Index.UpdateMessage(message, updatedFields)
			if err != nil {
				log.WithError(err).Warnf("failed to update message %s in index", hit.Id)
				erroneousMessages = append(erroneousMessages, userId.String())
			}
		}
	}
	return nil
}
