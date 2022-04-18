import type { Optional, Required } from 'simplytyped'
import createStore from './store'

export type DialogProps = {
  title?: string
  kind: 'createTranslation' | 'createCategory' | 'editCategory' | 'translation'
  parent?: string
  id?: string
}
export const state = createStore({
  initialValue: {
    showDeleted: false,
    serverStats: false,
    sidebarVisible: false,
    pageSize: 50,
    searchQuery: '',
    searchInTranslationValues: true,
    searchInTrasnaltions: true,
    searchInCategories: true,
    dialog: null as DialogProps | null,
    categorySortOn: 'key' as keyof ApiDef.Category,
    categorySortAsc: true,
    seenHints: {} as Record<string, [version: number, readAt: Date]>,
    collapse: {} as Record<string, boolean>,
    createOrganization: {} as ApiDef.OrganizationInput,
    createTranslation: {} as ApiDef.TranslationInput,
    createCategory: {} as ApiDef.CategoryInput,
    createProject: { locales: {} } as Required<ApiDef.ProjectInput, 'locales'>,
    projectSettings: {} as Record<string, { localeIds: string[] }>,
    openTranslationValueForm: '',
    createTranslationValue: {} as ApiDef.TranslationValueInput,
    toasts: {} as Record<
      string,
      Toast & { created: Date; _timeout: NodeJS.Timeout }
    >,
  },
  storage: {
    key: 'state',
    ignoreKeys: ['createOrganization', 'createTranslation', 'createTranslationValue', 'createProject', 'openTranslationValueForm']
  },
})

export function showDialog(d: DialogProps | null) {
  if (!d) { state.update(s => ({ ...s, dialog: null })) }
  state.update(s => {
    if (s.dialog) {
      return s
    }
    return { ...s, dialog: d }
  })
}

type Toast = {
  kind: 'error' | 'info' | 'warning'
  title: string
  message: string
  timeout: number
}

function hashCode(str: string) {
  let hash = 0
  for (let i = 0; i < str.length; ++i) {

    hash = Math.imul(31, hash) + str.charCodeAt(i)
  }

  return hash | 0
}

export function toast(t: Optional<Toast, 'timeout'>, key?: string) {
  let id = key || String(hashCode(t.kind + t.title + t.message))
  if (!t.timeout || t.timeout <= 0) {
    t.timeout = 8000
  }
  state.update((s) => {
    // If set previously, we clear the timeout, (and replace with our new one)
    const existing = s.toasts[id]
    if (existing?._timeout) {
      clearTimeout(existing._timeout)
    }
    return {
      ...s,
      toasts: {
        ...s.toasts,
        [id]: {
          ...existing,
          ...t,
          created: new Date(),
          _timeout: setTimeout(() => clearToast(id), t.timeout),
        },
      },
    }
  })
}

export function toastApiErr(err: ApiDef.APIError) {
  if (!err) {
    return
  }
  toast({
    kind: 'error',
    message: err.error?.code || '',
    title: err.error?.error || '',
  })
}

function checkToasts() {
  const now = new Date().getTime()
  state.update((s) => {
    const toasts = Object.entries(s.toasts)
    if (!toasts.length) {
      return s
    }
    return {
      ...s,
      toasts: toasts.reduce((r, [k, toast]) => {
        if (!toast?.created) {
          return r
        }
        const created = toast.created instanceof Date ? toast.created : new Date(toast.created)
        if (!created || isNaN(created.getTime())) {
          return r
        }
        if (created.getTime() + toast.timeout < now) {
          return r
        }
        r[k] = toast
        return r
      }, {}),
    }
  })
}
setTimeout(checkToasts, 1000)

export function clearToast(key: string) {
  state.update((s) => {
    const { [key]: _, ...toasts } = s.toasts
    if (!_) {
      return s
    }
    return {
      ...s,
      toasts,
    }
  })
}
