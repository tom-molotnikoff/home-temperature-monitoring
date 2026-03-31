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
    'dashboards',
    'cli-tool',
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
        'api/dashboards',
      ],
    },
    'configuration',
    {
      type: 'category',
      label: 'Development',
      items: [
        'development/architecture',
        'development/database',
        'development/code-patterns',
        'development/authentication',
        'development/websocket',
        'development/building-from-source',
        'development/docker-dev-environment',
        'development/testing',
        'development/releasing',
        'development/troubleshooting',
      ],
    },
  ],
};

export default sidebars;
