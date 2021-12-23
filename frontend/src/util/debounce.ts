import type { AnyFunc } from 'simplytyped'
// https://github.com/you-dont-need/You-Dont-Need-Lodash-Underscore#_debounce
function debounce<T extends AnyFunc>(func: T, wait: number, { leading }: { leading?: boolean }): T {
  let timeout: any;
  return function() {
    var context = this, args = arguments;
    clearTimeout(timeout);
    timeout = setTimeout(function() {
      timeout = null;
      if (!leading) func.apply(context, args);
    }, wait);
    if (leading && !timeout) func.apply(context, args);
  } as any;
}


export default debounce
