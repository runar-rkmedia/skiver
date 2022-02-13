import createStore from './store'
import type { Optional } from 'simplytyped'

export const state = createStore({
  initialValue: {
    showDeleted: false,
    serverStats: false,
    pageSize: 50,
    searchQuery: '',
    categorySortOn: 'key' as keyof ApiDef.Category,
    categorySortAsc: true,
    seenHints: {} as Record<string, [version: number, readAt: Date]>,
    collapse: {} as Record<string, boolean>,
    createOrganization: {} as ApiDef.OrganizationInput,
    createTranslation: {} as ApiDef.TranslationInput,
    createCategory: {} as ApiDef.CategoryInput,
    createProject: {} as ApiDef.ProjectInput,
    projectSettings: {} as Record<string, { localeIds: string[] }>,
    createTranslationValue: {} as ApiDef.TranslationValueInput,
    toasts: {} as Record<
      string,
      Toast & { created: Date; _timeout: NodeJS.Timeout }
    >,
  },
  storage: {
    key: 'state',
  },
})

type Toast = {
  kind: 'error' | 'info' | 'warning'
  title: string
  message: string
  timeout: number
}

function hashCode(s: string) {
  for (var i = 0, h = 0; i < s.length; i++)
    h = (Math.imul(31, h) + s.charCodeAt(i)) | 0
  return h
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
        if (toast.created.getTime() + toast.timeout < now) {
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
