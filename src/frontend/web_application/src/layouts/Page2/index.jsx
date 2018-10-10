import React, { Component } from 'react';
import PropTypes from 'prop-types';
import classnames from 'classnames';
import { Brand, Link } from '../../components/';
import { AppLoader } from '../../modules/appLoader';
import { BackgroundImage } from '../../modules/pi';
import { TabProvider } from '../../modules/tab';
import { PageActions } from '../../modules/control';
import { UserMenu } from '../../modules/user';
import { ScrollDetector } from '../../modules/scroll';
import Navigation from './components/Navigation';
import Footer from './components/Footer';
import NotificationCenter from './components/NotificationCenter';
import PageContainer from '../PageContainer';
import './style.scss';
import './header.scss';
import './navbar.scss';

// eslint-disable-next-line react/prefer-stateless-function
class Page extends Component {
  static propTypes = {
    children: PropTypes.node.isRequired,
  };

  render() {
    const { children } = this.props;

    return (
      <TabProvider>
        <BackgroundImage context="secure" className="l-page">
          {/* <div className="header-device">
            <PageContainer>
              Vous consultez actuellement vos messages dans un environement que vous avez classé
              comme <Link href="#">sûr</Link>, depuis un <Link href="#">appareil de
              confiance.</Link>
            </PageContainer>
          </div> */}
          <div className="l-header">
            <PageContainer>
              <Link to="/"><Brand className="l-header__brand" /></Link>
              {/* <div className="l-header__notif-menu"><Button href="#"><Icon type="bell"
              /></Button></div> */}
              <div className="l-header__user-menu">
                <UserMenu />
              </div>
            </PageContainer>
          </div>

          <PageContainer>
            <PageActions className="l-page__main-actions" />
          </PageContainer>

          <ScrollDetector
            offset={136}
            render={isSticky => (
              <div className="l-navbar">
                <div className={classnames('l-navbar__wrapper', { 'l-navbar__wrapper--sticky': isSticky })}>
                  <PageContainer>
                    <Navigation isSticky={isSticky} />
                  </PageContainer>
                </div>
              </div>
              )}
          />

          <PageContainer>
            {children}
          </PageContainer>

          <PageContainer>
            <Footer />
          </PageContainer>
        </BackgroundImage>
        <NotificationCenter />
      </TabProvider>
    );
  }
}

export default Page;
