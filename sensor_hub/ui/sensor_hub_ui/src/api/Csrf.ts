let csrfToken: string | null = null;

export function setCsrfToken(t: string | null) {
  csrfToken = t;
}

export function getCsrfToken(): string | null {
  return csrfToken;
}

