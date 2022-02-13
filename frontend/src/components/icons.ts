export const iconMap = {
  edit: 'fa-edit',
  code: 'fa-code',
  create: 'fa-plus',
  cancel: 'fa-ban',
  submit: 'fa-paper-plane',
  collapseUp: 'fa-angle-up',
  collapseDown: 'fa-angle-down',
  closeCross: 'fa-times',
  gEndpoint: 'fa-ethernet',
  gRequest: 'fa-network-wired',
  gStat: 'fa-chart-bar',
  gSchedule: 'fa-calendar',
  error: 'fa-exclamation-circle',
  warning: 'fa-exclamation-triangle',
  info: 'fa-info',
  success: 'fa-check',
  delete: 'fa-trash',
  copy: 'fa-copy',
  clock: 'fa-clock',
  refresh: 'fa-sync',
  play: 'fa-play',
  toggleOn: 'fa-toggle-on',
  toggleOff: 'fa-toggle-off',
  signIn: 'fa-sign-in-alt',
  loading: 'fad fa-spinner fa-spin'
} as const

export type Icon = keyof typeof iconMap
