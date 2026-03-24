---
id: user-management
title: User Management and RBAC
sidebar_position: 9
---

# User Management and RBAC

Sensor Hub includes a role-based access control (RBAC) system that governs what each user can see and do. Users are assigned roles, and roles contain permissions that control access to specific features and API endpoints.

## User accounts

Each user account has:

- A unique username
- An email address
- One or more roles
- An optional "must change password" flag

The initial admin user is created on first startup using the `SENSOR_HUB_INITIAL_ADMIN` environment variable (see [Installation](installation)). Additional users are created through the User Management page in the web UI or via the API.

User accounts can be disabled without deletion. A disabled user cannot log in but their data is retained.

## Creating and managing users

User management requires the `manage_users` permission. From the User Management page, administrators can:

- Create new user accounts with a username, email, password, and assigned roles
- Delete user accounts (you cannot delete your own account)
- Set the "must change password" flag on a user, which forces them to change their password on next login
- Assign or change a user's roles

While the "must change password" flag is active, the affected user can only access the password change page. All other pages and API endpoints return a 403 Forbidden response until the password is changed.

## Password management

- Users can change their own password from the web UI at any time
- Passwords are hashed using bcrypt with a configurable cost factor (default: 12)
- The bcrypt cost can be adjusted via the `auth.bcrypt.cost` property (see [Configuration Settings](configuration))

## Roles

Sensor Hub includes three built-in roles:

| Role   | Purpose                                        |
|--------|------------------------------------------------|
| admin  | Full administrative access to all features     |
| user   | Standard access for day-to-day use             |
| viewer | Read-only access to sensor data and dashboards |

A user's effective permissions are the union of all permissions granted to their assigned roles. A user can have multiple roles.

## Permissions

Permissions are granular access controls that are assigned to roles. Administrators can view and modify the permissions assigned to each role from the User Management page.

The following permissions are available:

| Permission             | Description                                     |
|------------------------|-------------------------------------------------|
| `manage_sensors`       | Add and edit sensors                            |
| `delete_sensors`       | Delete sensors                                  |
| `view_sensors`         | View the sensor list and sensor data            |
| `trigger_readings`     | Manually trigger sensor reading collection      |
| `view_alerts`          | View alert rules and alert history              |
| `manage_alerts`        | Create, edit, and delete alert rules            |
| `view_roles`           | View roles and their assigned permissions       |
| `manage_roles`         | Assign and remove permissions from roles        |
| `manage_users`         | Create, delete, and modify user accounts        |
| `view_users`           | View the user list                              |
| `view_readings`        | View temperature readings and charts            |
| `view_notifications`   | View notifications                              |
| `manage_notifications` | Dismiss notifications and configure preferences |
| `manage_properties`    | Update system configuration properties          |
| `view_properties`      | View system configuration properties            |
| `manage_oauth`         | Configure OAuth settings                        |

## Permission enforcement

Permissions are enforced at the API level. Each endpoint declares its required permission, and the server validates that the authenticated user's roles include that permission before processing the request. Requests without the required permission receive a 403 Forbidden response.

The web UI also uses permissions to control visibility. Navigation items, buttons, and page sections are shown or hidden based on the current user's permissions.
