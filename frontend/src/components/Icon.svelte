<script type="ts">
  import type { Colors } from '../appTypes'

  import { iconMap } from './icons'
  import type { Icon } from './icons'

  export let icon: Icon
  export let kind = 'fas'
  export let noMinWidth = false
  export let color: Colors = ''
  const colorMap = {
    error: 'color-error',
    warning: 'color-warning',
    info: 'color-info',
  }
</script>

<i
  class:minWidth={!noMinWidth}
  class={[
    color !== 'inherit' && (color ? `color-${color}` : colorMap[icon]),
    kind,
    iconMap[icon],
    $$props.class,
  ]
    .filter(Boolean)
    .join(' ')} />

<style>
  i {
    display: inline-block;
  }
  i.minWidth {
    min-width: 1.2em;
  }
</style>
