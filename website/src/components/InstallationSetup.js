import React, {useState} from 'react';
import {useLocation} from '@docusaurus/router';
import Tabs from '@theme/Tabs';
import styles from './InstallationSetup.module.css';

export function InstallationTabLabel({title, detail}) {
  return (
    <span className={styles.label}>
      {title}
      <span className={styles.detail}>{detail}</span>
    </span>
  );
}

export function InstallationMethodTabs({children}) {
  return (
    <Tabs className={styles.tabs} queryString="method" defaultValue="script">
      {children}
    </Tabs>
  );
}

function useInstallationMethod() {
  const {search} = useLocation();
  const params = new URLSearchParams(search);
  return {
    method: params.get('method') || 'script',
    platform: params.get('platform'),
  };
}

export default function InstallationSetup({children}) {
  const {method} = useInstallationMethod();

  if (method === 'github-actions') {
    return <p>The GitHub Actions installer configures PATH for your workflow. No shell configuration or new-terminal check is needed.</p>;
  }

  if (method === 'others') {
    return <p>Follow the CircleCI Orb or Dev Container Feature setup instructions linked in step 1. The local terminal instructions do not apply to those environments.</p>;
  }

  return children;
}

export function InstallationShellTabs({children}) {
  const {method, platform} = useInstallationMethod();
  const [selection, setSelection] = useState({method, urlPlatform: platform, platform});
  if (selection.method !== method || selection.urlPlatform !== platform) {
    // A new method resets the suggestion unless the link also selects a platform.
    setSelection({
      method,
      urlPlatform: platform,
      platform: selection.urlPlatform !== platform ? platform : null,
    });
  }
  const suggestedPlatforms = {
    script: 'linux',
    homebrew: 'macos',
    winget: 'windows',
    scoop: 'windows',
    go: 'linux',
    binary: 'linux',
  };
  const defaultValue = ['linux', 'macos', 'windows', 'command-prompt'].includes(selection.platform)
    ? selection.platform
    : suggestedPlatforms[method] || suggestedPlatforms.script;

  return (
    <Tabs className={styles.tabs} key={`${method}:${platform}`} defaultValue={defaultValue}>
      {children}
    </Tabs>
  );
}
