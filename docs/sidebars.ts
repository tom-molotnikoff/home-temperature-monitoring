import type {SidebarsConfig} from '@docusaurus/plugin-content-docs';

const sidebars: SidebarsConfig = {
  docs: [
    'overview',
    'prerequisites',
    'installation',
    'nginx-setup',
    'upgrading',
    'uninstalling',
    {
      type: 'category',
      label: 'Sensors',
      link: { type: 'doc', id: 'sensors/sensors' },
      items: [
        'sensors/http-temperature',
        'sensors/zigbee',
        'sensors/managing-sensors-ref',
      ],
    },
    'alerts-and-notifications',
    'session-management',
    'user-management',
    'dashboards',
    {
      type: 'category',
      label: 'How-to Guides',
      items: [
        'how-to/connect-http-sensor',
        'how-to/connect-zigbee-device',
        'how-to/monitor-energy-usage',
      ],
    },
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
        'development/sensor-drivers',
        'development/mqtt',
        'development/configuration-package',
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
