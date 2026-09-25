import React, {Children, useEffect, useRef} from 'react';
import {useHistory, useLocation} from '@docusaurus/router';
import useIsBrowser from '@docusaurus/useIsBrowser';
import Link from '@docusaurus/Link';
import styles from './InstallationSetup.module.css';

const installationMethods = {
  script: {platform: 'linux'},
  homebrew: {platform: 'macos'},
  winget: {platform: 'windows'},
  scoop: {platform: 'windows'},
  go: {platform: 'linux'},
  binary: {platform: 'linux'},
  'github-actions': {message: 'The GitHub Actions installer configures PATH for your workflow. No local shell setup or new-terminal check is needed.'},
  circleci: {message: 'Follow the CircleCI Orb instructions in step 1. No local shell setup or new-terminal check is needed.'},
  devcontainer: {message: 'Follow the Dev Container Feature instructions in step 1. No local shell setup or new-terminal check is needed.'},
};

function getInstallationMethod(search) {
  const method = new URLSearchParams(search).get('method');
  return Object.hasOwn(installationMethods, method) ? installationMethods[method] : installationMethods.script;
}

function useInstallationLocation() {
  const location = useLocation();
  const isBrowser = useIsBrowser();
  // The server and the first client render must use the same defaults.
  return {...location, search: isBrowser ? location.search : '', hash: isBrowser ? location.hash : ''};
}

export function InstallationTabLabel({title, detail}) {
  return <span className={styles.label}>{title}<span className={styles.detail}>{detail}</span></span>;
}

export function InstallationShellLink({platform, children, hash}) {
  const location = useInstallationLocation();
  const params = new URLSearchParams(location.search);
  params.set('platform', platform);
  const targetHash = getInstallationMethod(location.search).message ? '#2-set-the-environment-variable-path' : hash;
  return <Link to={`${location.pathname}?${params}${targetHash}`}>{children}</Link>;
}

function InstallationTabs({children, parameter, fallback, label}) {
  const location = useInstallationLocation();
  const history = useHistory();
  const tabList = useRef(null);
  const items = Children.toArray(children);
  const params = new URLSearchParams(location.search);
  const requested = params.get(parameter);
  const selected = items.some(({props}) => props.value === requested) ? requested : fallback;

  function select(value) {
    if (value === selected) return;
    const top = tabList.current.getBoundingClientRect().top;
    params.set(parameter, value);
    if (parameter === 'method') params.delete('platform');
    history.push({...location, search: `?${params}`, hash: ''});
    // Clearing an old anchor must not send the reader back to the page top.
    requestAnimationFrame(() => {
      if (tabList.current) window.scrollBy(0, tabList.current.getBoundingClientRect().top - top);
    });
  }

  function onKeyDown(event, index) {
    const offsets = {ArrowRight: 1, ArrowLeft: -1, Home: -index, End: items.length - 1 - index};
    if (!(event.key in offsets)) return;
    event.preventDefault();
    const next = (index + offsets[event.key] + items.length) % items.length;
    event.currentTarget.parentElement.children[next].focus();
    select(items[next].props.value);
  }

  return <div className="tabs-container">
    <div ref={tabList} className={`tabs ${styles.tabs}`} role="tablist" aria-label={label}>
      {items.map(({props}, index) => <button
        type="button" role="tab" key={props.value}
        id={`${parameter}-tab-${props.value}`} aria-controls={`${parameter}-panel-${props.value}`}
        aria-selected={selected === props.value} tabIndex={selected === props.value ? 0 : -1}
        className="tabs__item" onClick={() => select(props.value)} onKeyDown={(event) => onKeyDown(event, index)}>
        {props.label}
      </button>)}
    </div>
    {items.map(({props}) => <div key={props.value} role="tabpanel"
      id={`${parameter}-panel-${props.value}`} aria-labelledby={`${parameter}-tab-${props.value}`}
      data-installation-parameter={parameter} data-installation-value={props.value}
      hidden={selected !== props.value} className="margin-top--md">
      {props.children}
    </div>)}
  </div>;
}

export function InstallationMethodTabs({children}) {
  const location = useInstallationLocation();
  const history = useHistory();
  useEffect(() => {
    let cancelDetailsScroll = () => {};
    // Hidden headings cannot be native scroll targets. Reveal their panel first,
    // including on a full page load when Docusaurus does not scroll again.
    function reveal(hash) {
      let id;
      try { id = decodeURIComponent(hash.slice(1)); } catch { return; }
      const target = document.getElementById(id);
      if (!target) return;
      const panel = target.closest('[data-installation-parameter]');
      const params = new URLSearchParams(location.search);
      if (panel && params.get(panel.dataset.installationParameter) !== panel.dataset.installationValue) {
        params.set(panel.dataset.installationParameter, panel.dataset.installationValue);
        if (panel.dataset.installationParameter === 'method') params.delete('platform');
        history.replace({...location, search: `?${params}`, hash});
        return;
      }
      const details = target.closest('details[data-collapsed="true"]');
      if (details) {
        // Docusaurus tracks collapse state separately from the native open
        // attribute. Use its toggle and wait until the content is expanded.
        cancelDetailsScroll();
        function onExpanded(event) {
          if (event.propertyName !== 'height' || event.target.parentElement !== details) return;
          cancelDetailsScroll();
          reveal(hash);
        }
        details.addEventListener('transitionend', onExpanded);
        cancelDetailsScroll = () => details.removeEventListener('transitionend', onExpanded);
        details.querySelector('summary').click();
        return;
      }
      target.scrollIntoView();
    }
    const frame = requestAnimationFrame(() => reveal(location.hash));
    function onClick(event) {
      if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return;
      const anchor = event.target.closest('a[href]');
      if (!anchor) return;
      const url = new URL(anchor.href);
      if (url.origin !== window.location.origin || url.pathname !== location.pathname || !url.hash) return;
      // Repeated clicks on an unchanged hash must reveal and scroll too.
      if (url.search === location.search && url.hash === location.hash) reveal(url.hash);
    }
    // Details stops bubbling clicks, including clicks on heading permalinks.
    document.addEventListener('click', onClick, true);
    return () => {
      cancelAnimationFrame(frame);
      cancelDetailsScroll();
      document.removeEventListener('click', onClick, true);
    };
  }, [location.search, location.hash, location.pathname, history]);
  return <InstallationTabs parameter="method" fallback="script" label="Installation method">{children}</InstallationTabs>;
}

export default function InstallationSetup({children}) {
  const {search} = useInstallationLocation();
  const {message} = getInstallationMethod(search);
  return message ? <p>{message}</p> : children;
}

export function InstallationShellTabs({children}) {
  const {search} = useInstallationLocation();
  const {platform} = getInstallationMethod(search);
  return <InstallationTabs parameter="platform" fallback={platform || 'linux'} label="Shell">{children}</InstallationTabs>;
}
