/**
 * Scrolls to the element with the given ID and briefly highlights it
 * with a pulse animation to draw the user's attention.
 */
export function scrollToAndHighlight(elementId: string) {
  const el = document.getElementById(elementId);
  if (!el) return;

  el.scrollIntoView({ behavior: 'smooth', block: 'center' });

  el.style.transition = 'box-shadow 0.3s ease';
  el.style.boxShadow = '0 0 0 3px var(--mui-palette-primary-main, #1976d2)';

  setTimeout(() => {
    el.style.boxShadow = '';
    setTimeout(() => { el.style.transition = ''; }, 300);
  }, 1500);
}
