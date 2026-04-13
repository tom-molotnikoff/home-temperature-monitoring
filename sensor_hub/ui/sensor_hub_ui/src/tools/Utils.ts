import type {AuthUser} from "../providers/AuthContext.tsx";

export const hasPerm = (user: AuthUser | undefined, perm: string) => {
  if (!user) return false;
  if (user.permissions && user.permissions.includes(perm)) return true;
  // fallback to admin role for compatibility
  if (user.roles && user.roles.includes('admin')) return true;
  return false;
}

/**
 * Parse a reading time string from the API (stored as UTC without a Z suffix)
 * into a Date that is correctly interpreted as UTC.
 */
export function parseUTCTime(time: string): Date {
  const iso = time.includes("T") ? time : time.replace(" ", "T");
  return new Date(iso.endsWith("Z") ? iso : iso + "Z");
}