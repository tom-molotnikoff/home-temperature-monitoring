import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    'overview',
    'prerequisites',
    'installation',
    'nginx-setup',
    'upgrading',
    'uninstalling',
    'deploying-sensors',
    'managing-sensors',
    'alerts-and-notifications',
    'session-management',
    'user-management',
    'cli',
    'llm-skills',
    'telemetry',
    {
      type: 'category',
      label: 'API Reference',
      items: [
        'api/authentication',
        'api/sensors-and-readings',
        'api/alerts-and-notifications',
        'api/users-roles-sessions',
        'api/properties-and-oauth',
        'api/api-keys',
      ],
    },
    'configuration',
    {
      type: 'category',
      label: 'Development',
      items: [
        'development/building-from-source',
        'development/docker-dev-environment',
        'development/testing',
        'development/releasing',
      ],
    },
  ],
};

export default sidebars;
