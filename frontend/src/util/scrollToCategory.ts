
export function scrollToCategory(e) {
  const el = document.getElementById(
    e.target?.getAttribute('href')?.replace('#', '') ||
    (
      'cat-' +
      e.target?.getAttribute?.('data-key')
    )
  )
  if (!el) {
    return
  }
  el.scrollIntoView({ behavior: 'auto' })
  // The categories we are scolling to are lazy-loaded, so they will move areound a bit as we scroll
  // This hopefully mitigates this, but it is very hacky
  setTimeout(() => el?.scrollIntoView({ behavior: 'smooth' }), 100)
  setTimeout(() => el?.scrollIntoView({ behavior: 'smooth' }), 200)
  setTimeout(() => el?.scrollIntoView({ behavior: 'smooth' }), 300)
  setTimeout(() => el?.scrollIntoView({ behavior: 'smooth' }), 400)
  return el
}
