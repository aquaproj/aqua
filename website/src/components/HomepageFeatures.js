import React from 'react';
import clsx from 'clsx';
import styles from './HomepageFeatures.module.css';

const FeatureList = [
  {
    title: 'Manage CLI declaratively',
    description: (
      <>
        <img width="600" alt="image" src="https://user-images.githubusercontent.com/13323303/176584746-3ec582d7-9cbe-41fe-ab1a-1b26e1436518.png" />
      </>
    ),
  },
  {
    title: 'Install tools easily',
    description: (
      <>
        <img width="600" alt="image" src="https://user-images.githubusercontent.com/13323303/176585183-b8616482-5e3b-4f99-be98-6e7d752c5dbc.png" />
      </>
    ),
  },
  {
    title: 'Lazy Install',
    description: (
      <>
        <img width="600" alt="image" src="https://user-images.githubusercontent.com/13323303/176579462-e843b334-86a0-4b16-88ab-e39b8fbfa0a9.png" />
        aqua installs a tool automatically when the tool is executed.
      </>
    ),
  },
  {
    title: 'Change tool versions per project',
    description: (
      <>
        <img width="600" alt="image" src="https://user-images.githubusercontent.com/13323303/176584607-951eefc0-572b-48eb-84f7-b2c17efb5cd4.png" />
        aqua manages tool versions per configuration file.
        You can install multiple versions and switch them seamlessly.
      </>
    ),
  },
  {
    title: 'Interactive Search',
    description: (
      <>
        <img width="600" alt="image" src="https://user-images.githubusercontent.com/13323303/176598695-134f0ad4-296b-4491-a3da-e5e454afdf4b.png" />
      </>
    ),
  },
  {
    title: 'Renovate Integration',
    description: (
      <>
        <img width="600" alt="image" src="https://user-images.githubusercontent.com/13323303/176582627-44f27c48-213b-44da-b18f-d4d482ef2f56.png" />

        aqua provides <a href="/docs/products/aqua-renovate-config/">Renovate Config Preset</a>, so you can update tools by Renovate easily.
      </>
    ),
  },
  {
    title: 'GitHub Actions & CircleCI Orb',
    description: (
      <>
        <img width="600" alt="image" src="https://user-images.githubusercontent.com/13323303/176584418-c6a4adca-e4d8-45aa-98b0-6ae3b543b007.png" />
        aqua has GitHub Actions and CircleCI Orb to install aqua and update checksum files.
        Please see <a href="/docs/products/aqua-installer">aqua-installer</a> and <a href="/docs/products/circleci-orb-aqua">circleci-orb-aqua</a>
      </>
    ),
  },
  {
    title: 'Secure',
    description: (
      <>
        aqua installs tools securely.
        aqua supports <a href="/docs/reference/security/checksum/">Checksum Verification</a>, <a href="/docs/reference/security/policy-as-code/">Policy as Code</a>, <a href="/docs/reference/security/cosign-slsa/">Cosign and SLSA Provenance</a>, <a href="/docs/reference/security/github-artifact-attestations">GitHub Artifact Attestations</a>, and <a href="/docs/reference/security/minisign/">Minisign</a>.
        Please see <a href="/docs/reference/security">Security</a>.
      </>
    ),
  },
  {
    title: 'Single Binary / Cross Platform',
    description: (
      <>
        aqua works as a single binary, and basically aqua doesn't depend on anything.
        aqua supports Windows, macOS, and Linux.
        aqua can be used for both local development and CI. You can manage CLI in the unified way.
      </>
    ),
  },
];

function Feature({ title, description }) {
  return (
    <div className={clsx('col col--6')}>
      <div className="text--center padding-horiz--md">
        <h3>{title}</h3>
        <p>{description}</p>
      </div>
    </div>
  );
}

export default function HomepageFeatures() {
  return (
    <section className={styles.features}>
      <div className="container">
        <div className="row">
          {FeatureList.map((props, idx) => (
            <Feature key={idx} {...props} />
          ))}
        </div>
      </div>
    </section>
  );
}
