// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require('prism-react-renderer/themes/github');
const darkCodeTheme = require('prism-react-renderer/themes/dracula');

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'aqua',
  tagline: 'Declarative CLI Version Manager. Unify tool versions in teams, projects, and CI. Easy, painless, and secure.',
  url: 'https://aquaproj.github.io',
  baseUrl: '/',
  onBrokenLinks: 'throw',
  onBrokenMarkdownLinks: 'warn',
  favicon: 'https://raw.githubusercontent.com/aquaproj/aqua/main/logo/aqua_without_text.svg',
  organizationName: 'aquaproj', // Usually your GitHub org/user name.
  projectName: 'aqua', // Usually your repo name.

  presets: [
    [
      '@docusaurus/preset-classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve('./sidebars.js'),
          // Please change this to your repo.
          editUrl: 'https://github.com/aquaproj/aquaproj.github.io/edit/main',
        },
        theme: {
          customCss: require.resolve('./src/css/custom.css'),
        },
        blog: {
          routeBasePath: '/blog',
          feedOptions: {
            type: 'all',
            copyright: `Copyright © 2021 Shunsuke Suzuki. Built with Docusaurus.`,
          },
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      announcementBar: {
        id: 'security-keyring',
        content:
          '<a href="/docs/reference/security/keyring">Manage a GitHub access token using Keyring (2025-05-06)</a>',
        backgroundColor: '#7FFF00',
        textColor: '#091E42',
        isCloseable: true,
      },
      navbar: {
        title: 'aqua',
        logo: {
          alt: 'aqua Logo',
          src: 'https://raw.githubusercontent.com/aquaproj/aqua/main/logo/aqua_without_text.svg',
        },
        items: [
          {
            type: 'doc',
            docId: 'index',
            position: 'left',
            label: 'Docs',
          },
          {
            href: '/docs/#contact-us',
            label: 'Contact',
            position: 'right',
          },
          {
            href: 'https://github.com/aquaproj/aqua',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              {
                label: 'Overview',
                to: '/docs/',
              },
              {
                to: '/docs/tutorial',
                label: 'Tutorial',
              },
              {
                to: '/docs/install',
                label: 'Install',
              },
              {
                to: '/docs/reference/slide-blog',
                label: 'Blog',
              },
            ],
          },
          {
            title: 'More',
            items: [
              {
                href: 'https://github.com/orgs/aquaproj/discussions',
                label: 'Question',
              },
              {
                label: 'GitHub',
                href: 'https://github.com/aquaproj/aqua',
              },
              {
                href: 'https://github.com/aquaproj/aqua-registry',
                label: 'Standard Registry',
              },
              {
                href: 'https://github.com/sponsors/aquaproj',
                label: 'Sponsor',
              },
              {
                href: 'https://github.com/aquaproj/aqua/releases',
                label: 'Changelog',
              },
              {
                href: 'https://asciinema.org/a/498262?autoplay=1',
                label: 'Demo',
              },
              {
                label: 'X (Formerly Twitter)',
                href: 'https://x.com/aquaclivm',
              },
              {
                label: 'Discord',
                href: 'https://discord.gg/CBGesz8yBX'
              },
            ],
          },
        ],
        copyright: `Copyright © 2021 Shunsuke Suzuki. Built with Docusaurus.`,
      },
      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
      },
      algolia: {
        appId: 'OBO5AZ814M',
        // Public API key: it is safe to commit it
        apiKey: '7ebd7217e9bd4836c5094b4acdf1a0c9',
        indexName: 'aquaproj',
        // Optional: see doc section below
        // contextualSearch: true,
        // Optional: Specify domains where the navigation should occur through window.location instead on history.push. Useful when our Algolia config crawls multiple documentation sites and we want to navigate with window.location.href to them.
        // externalUrlRegex: 'external\\.com|domain\\.com',

        // Optional: Algolia search parameters
        searchParameters: {},

        //... other Algolia params
      },
    }),
};

module.exports = config;
