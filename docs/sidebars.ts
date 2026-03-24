import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    'overview',
    'prerequisites',
    'installation',
    'upgrading',
    'deploying-sensors',
    'managing-sensors',
    'alerts-and-notifications',
    'session-management',
    'user-management',
    {
      type: 'category',
      label: 'API Reference',
      items: [
        'api/authentication',
        'api/sensors-and-readings',
        'api/alerts-and-notifications',
        'api/users-roles-sessions',
        'api/properties-and-oauth',
      ],
    },
    'configuration',
  ],
};

export default sidebars;
