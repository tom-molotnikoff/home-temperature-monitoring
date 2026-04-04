import {themes as prismThemes} from 'prism-react-renderer';
import type {Config} from '@docusaurus/types';
import type * as Preset from '@docusaurus/preset-classic';

const config: Config = {
  title: 'Sensor Hub',
  tagline: 'Home Temperature Monitoring System',
  favicon: 'img/favicon.ico',

  future: {
    v4: true,
  },

  url: 'https://sensor-hub.docs',
  baseUrl: '/docs/',

  onBrokenLinks: 'throw',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      {
        docs: {
          sidebarPath: './sidebars.ts',
          routeBasePath: '/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      } satisfies Preset.Options,
    ],
  ],

  themeConfig: {
    colorMode: {
      respectPrefersColorScheme: true,
    },
    navbar: {
      title: 'Sensor Hub Documentation',
      logo: {
        alt: 'Sensor Hub',
        src: 'img/favicon.svg',
      },
      items: [
        {
          type: 'docSidebar',
          sidebarId: 'docs',
          position: 'left',
          label: 'User Guide',
        },
      ],
    },
    footer: {
      style: 'dark',
      copyright: `Copyright \u00a9 ${new Date().getFullYear()} Tom Molotnikoff.`,
    },
    prism: {
      theme: prismThemes.github,
      darkTheme: prismThemes.dracula,
      additionalLanguages: ['bash', 'json', 'properties'],
    },
  } satisfies Preset.ThemeConfig,
};

export default config;
