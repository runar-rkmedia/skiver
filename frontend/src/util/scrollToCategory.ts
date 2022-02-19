
export const scrollToCategory = (e: any) => {
  const el = document.getElementById(
    e.currentTarget?.getAttribute('href')?.replace('#', '') ||
    (
      'cat-' +
      e.currentTarget?.getAttribute?.('data-key')
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

export const createCategoryAnchorProps = (c: { key?: string }) => {
  const key = c.key || "_root_"

  return {

    href: '#cat-' + key,
    'data-key': key
  }
}
