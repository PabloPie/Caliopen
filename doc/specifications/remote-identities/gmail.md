# Gmail

This describes the  `SASL XOAUTH2 Mechanism` used by gmail to access to mailbox.
Authenticate with [Oauth 2.0](./oauth.md)
Then use the Bearer to initialize SMTP/IMAP operations

## The url for the request credentials popup

```
https://accounts.google.com/o/oauth2/v2/auth?
  scope=https://mail.google.com/&
  access_type=offline&
  include_granted_scopes=true&
  state=state_parameter_passthrough_value&
  redirect_uri=<caliopen instance/api/v2/oauth/google/callback>
  response_type=code&
  client_id=<client_id>
```

## How to configure the gmail protocol

* Go to https://console.developers.google.com/
* Create a new project (choose a name for the project)
* go to "credentials" page
  * click on "create credentials"
  * select "OAuth client ID"
  * choose application type: "Web application"
  * name it as you want (for example: "oauth <my> caliopen instance")
  *  **FIXME** in the field "Authorized redirect URIs" enter the url for this api: `https://<caliopen instance/api/v2/oauth/google/callback>` then press "Enter"
  * Click "Create"
* on credentials page, copy the client_id of the credential you have created
* **FIXME** and add it to _the config_


More information on  https://developers.google.com/identity/protocols/OAuth2UserAgent
