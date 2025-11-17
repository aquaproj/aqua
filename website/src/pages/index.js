import React from 'react';
import clsx from 'clsx';
import Layout from '@theme/Layout';
import Link from '@docusaurus/Link';
import useDocusaurusContext from '@docusaurus/useDocusaurusContext';
import styles from './index.module.css';
import HomepageFeatures from '../components/HomepageFeatures';

function HomepageHeader() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <header className={clsx('hero hero--primary', styles.heroBanner)}>
      <div className="container">
        <img src="https://raw.githubusercontent.com/aquaproj/aqua/main/logo/aqua_without_text.svg" alt="aqua Logo" className="top__logo" />
        <h1 className="hero__title">{siteConfig.title}</h1>
        <p className="hero__subtitle">{siteConfig.tagline}</p>
        <div className={styles.buttons}>
          <Link
            className="button button--secondary button--lg"
            to="https://asciinema.org/a/498262?autoplay=1">
            Demo
          </Link>
          <Link
            className="button button--secondary button--lg"
            to="/docs/tutorial">
            Quick Start
          </Link>
          <Link
            className="button button--secondary button--lg"
            to="https://notebooklm.google.com/notebook/874e89e4-66a1-459a-82c9-923b81501a71">
            NotebookLM
          </Link>
        </div>
      </div>
    </header>
  );
}

export default function Home() {
  const {siteConfig} = useDocusaurusContext();
  return (
    <Layout
      title={`aqua Official Website`}
      description="aqua Official Website">
      <HomepageHeader />
      <main>
        <HomepageFeatures />
      </main>
    </Layout>
  );
}
