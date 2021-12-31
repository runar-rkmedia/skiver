import type { AnyFunc } from 'simplytyped'
// https://github.com/you-dont-need/You-Dont-Need-Lodash-Underscore#_debounce
function debounce<T extends AnyFunc>(
  func: T,
  wait: number,
  { leading }: { leading?: boolean }
): T {
  let timeout: any
  return function () {
    var context = this,
      args = arguments
    clearTimeout(timeout)
    let called = false
    if (leading && !timeout) {
      func.apply(context, args)
      called = true
    }
    timeout = setTimeout(function () {
      timeout = null
      if (called) {
        return
      }
      func.apply(context, args)
    }, wait)
  } as any
}

export default debounce
