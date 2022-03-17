import { fetchApi, methods, wsSubscribe } from './apiFetcher'
import type { ApiFetchOptions } from './apiFetcher'
import createStore from './store'
import type { AnyFunc } from 'simplytyped'
import { parseDate } from 'dates'
import { isPast } from 'date-fns'
import { derived, get } from 'svelte/store'
import { toast, toastApiErr } from 'state'
export function objectKeys<T extends object>(obj: T) {
  return Object.keys(obj) as Array<keyof T>
}

/**
 * Typedefintions are created by running `yarn gen`.
 *
 * This will use the generated swagger defintions from Gobyoall-api. (which again are created from running go generate)
 */

export type DB = {
  missingTranslation: Record<string, ApiDef.MissingTranslation>
  project: Record<string, ApiDef.Project>
  organization: Record<string, ApiDef.Organization>
  category: Record<string, ApiDef.Category>
  translationValue: Record<string, ApiDef.TranslationValue>
  locale: Record<string, ApiDef.Locale>
  translation: Record<string, ApiDef.Translation>
  login: ApiDef.LoginResponse
  serverInfo: ApiDef.ServerInfo
  responseStates: Omit<Record<keyof DB, { loading: boolean; error?: ApiDef.APIError }>, 'responseStates' | 'serverInfo'>
}


export const api = {
  join: {
    get: (id: string) => fetchApi<ApiDef.Organization>('join/' + id, () => null, { method: 'GET' }),
    post: (id: string, payload: ApiDef.JoinInput) => fetchApi<ApiDef.LoginResponse>('join/' + id, () => null, { method: 'POST', body: payload }),
  },
  serverInfo: (options?: ApiFetchOptions) =>
    fetchApi<ApiDef.ServerInfo>(
      'serverInfo',
      (e) => db.update((s) => ({ ...s, serverInfo: e })),
      options
    ),
  translation: CrudFactory<ApiDef.TranslationInput, 'translation',
    ApiDef.UpdateTranslationInput
  >(
    'translation'
  ),
  missingTranslation: {
    list: apiGetListFactory<'missingTranslation'>('missing', 'missingTranslation')
  },
  project: CrudFactory<ApiDef.ProjectInput, 'project', ApiDef.UpdateProjectInput>('project'),
  snapshotMeta: {
    create: apiCreateFactory<ApiDef.CreateSnapshotInput, 'project'>('project/snapshot', 'project'),
  },
  organization: CrudFactory<ApiDef.OrganizationInput, 'organization'>('organization'),
  category: CrudFactory<ApiDef.CategoryInput, 'category'>('category'),
  translationValue: CrudFactory<
    ApiDef.TranslationValueInput,
    'translationValue',
    ApiDef.UpdateTranslationValueInput
  >('translationValue'),
  locale: CrudFactory<ApiDef.LocaleInput, 'locale'>('locale'),
  logout: () => fetchApi<ApiDef.LogoutResponse>('logout', () => {
    return db.update(({ login, ...s }) => {
      login.ok = false
      windowPost.loginStatus(false)
      localStorage.setItem('login-response', JSON.stringify(login))

      return { ...s, login }
    })
  }, { method: "POST" }),

  login: {
    get: () =>
      fetchApi<ApiDef.LoginResponse>(
        'login',
        (login) => {
          localStorage.setItem('login-response', JSON.stringify(login))
          windowPost.loginStatus(!!login.ok)
          return db.update((s) => ({ ...s, login }))
        },
        { method: 'GET' }
      ),
    post: async (args: ApiDef.LoginInput) => {
      db.update((s) => ({
        ...s,
        responseStates: {
          ...s.responseStates,
          login: {
            ...s.responseStates.login,
            loading: true,
          },
        },
      }))
      const [res, err] = await fetchApi<ApiDef.LoginResponse>(
        'login',
        (login) => {
          localStorage.setItem('login-response', JSON.stringify(login))
          return db.update((s) => ({ ...s, login }))
        },
        { method: 'POST', body: args }
      )
      windowPost.loginStatus(!!res?.data?.ok)
      db.update((s) => ({
        ...s,
        responseStates: {
          ...s.responseStates,
          login: {
            loading: false,
            ...res,
            ...(!!err && { error: err }),
          },
        },
      }))
      return [res, err]
    },
  },
} as const

const tryOrNull = <T extends AnyFunc>(
  f: T,
  onErr?: (err: Error) => void
): null | ReturnType<T> => {
  try {
    return f()
  } catch (err) {
    onErr?.(err)
    return null
  }
}

const tryJsonParse = <T extends any>(s: string | null): null | T => {
  if (!s) {
    return null
  }
  return tryOrNull(() => JSON.parse(s))
}

const initialLoginResponse = (): ApiDef.LoginResponse => {
  const r: ApiDef.LoginResponse = {
    ok: false,
    username: '',
    created_at: '',
    updated_by: '',
    id: '',
  }

  const l = tryJsonParse<ApiDef.LoginResponse>(
    localStorage.getItem('login-response')
  )
  if (!l) {
    return r
  }
  if (!l.ok) {
    return l
  }
  l.ok = false
  if (!l.expires) {
    return l
  }
  const ex = parseDate(l.expires)
  if (!ex || isPast(ex)) {
    return l
  }

  l.ok = true
  return l
}

export const db = createStore<DB, null>({
  initialValue: objectKeys(api).reduce(
    (r, k) => {
      switch (k) {
        case 'login':
          r[k] = initialLoginResponse()
          break
        default:
          r[k] = {}
      }
      return r
    },
    {
      responseStates: objectKeys(api).reduce(
        (r, k) => ({ ...r, [k]: { loading: false } }),
        {}
      ),
    } as DB
  ),
})

export const projectCategoriesByKeyLength = derived(db, ($db) => {
  const sorted = Object.values($db.category)
    .reduce((r, c) => {
      if (!c.project_id) {
        return r
      }
      const length = getKeyLength(c.key)
      if (!r[c.project_id]) {
        r[c.project_id] = []
      }
      if (!r[c.project_id][length]) {
        r[c.project_id][length] = [c]
        return r
      }
      r[c.project_id][length].push(c)
      return r
    }, {})

  return sorted
})




const getKeyLength = (key?: string) => {
  if (!key) {
    return 0
  }
  return key?.split('.').length
}

wsSubscribe({
  onMessage: (msg) => {
    // only used in development
    if (msg.kind === 'dist') {
      window.location.reload()
      return
    }
    if (!msg.contents) {
      return
    }
    if (typeof msg.contents !== 'object') {
      return
    }
    if (!msg.contents?.id) {
      console.warn("received message without id", msg)
      return
    }

    replaceField(msg.kind, msg.contents as any, msg.contents.id, msg.variant)
  },
  autoReconnect: true,
})

const windowPost = {
  translationMessage: (variant: 'create' | 'update' | 'soft-delete', tv: ApiDef.TranslationValue, store: DB, extra?: any) => {
    if (!isWithinIframe()) {
      return
    }
    const t = !!tv && store.translation[tv.translation_id || '']
    const locale = !!tv && store.locale[tv.locale_id || '']
    const category = store.category[t?.category || '']
    const key = (!!category && !!t) ? [category.key].join('.') + "." + t.key : ''
    const text = `TranslationValue"-${variant}: ${key} is now '${tv.value}' for locale ${locale?.title}`

    const msg = {
      extra, translationValue: tv, translation: t, locale, category, text, key, value: tv.value, kind: 'translation-value-change', variant
    }
    switch (true) {
      case !(tv as any):
      case !(t as any):
      case !(category as any):
      case !(variant as any):
      case !(store as any):
      case !(key as any):
        console.warn("[windowPost.translationMessage]: Message has missing arguments:", { msg, store, extra })
    }
    console.debug("[windowPost.translationMessage]: Posting message:", msg)
    window.parent.postMessage(msg, "*")

  },
  error: (msg: string, details?: any) => {
    if (!isWithinIframe()) {
      return
    }
    console.error(msg, details)
    window.parent.postMessage({ kind: 'error', msg, details }, "*")
  },
  loginStatus: (ok: boolean) => {
    if (!isWithinIframe()) {
      return
    }

    window.parent.postMessage({ kind: 'login-status', ok }, "*")
  }

}

const isWithinIframe = () => {
  return window.self !== window.top
}

/** Messages from parent when webapp is loaded via an iframe */
function handleMsg(e: MessageEvent) {
  console.log('login-status')
  if (!e.data) {
    return
  }
  const kind = e.data?.kind
  switch (kind) {
    case 'login-status':
      {

        // some clients which is skiver as an iframe may request this information on an interval, like every second.
        // This may be a bit too frequent, but skiver should still make this fast.
        // If this gets slow, skiver should optimize it through any means necessary
        const store = get(db)
        windowPost.loginStatus(!!store.login.ok)
      }
      return
    case 'post-edit':
      const { context, category, locale, translation, value, project } = e.data as { category?: string, locale?: string, translation?: string, value?: string, project?: string, context?: string }
      if (!category || !locale || !translation || !value || !project) {
        windowPost.error("post-edit message had some missing arguments", e.data)
        return
      }
      const store = get(db)
      const p = Object.values(store.project).find(c => c.id === project || c.short_name === project)
      if (!p) {
        windowPost.error("post-edit failed to find the project: ", project)
        return
      }
      const cat = Object.values(store.category).find(c => c.project_id === p.id && c.key === category)
      if (!cat) {
        windowPost.error("post-edit failed to find the category: ", { category, project: p })
        return
      }
      let t: ApiDef.Translation | null = null
      loop: for (const tid of cat.translation_ids || []) {
        const found = store.translation[tid]
        if (found && found.key === translation) {

          t = found
          break loop
        }

      }

      if (!t) {
        windowPost.error("post-edit failed to find the translation: ", { cateory: cat, translations: cat.translation_ids?.map(tid => store.translation[tid]) })
        return
      }
      const loc = Object.values(store.locale).find(c => {
        if (locale === c.ietf) {
          return true
        }
        if (locale === c.iso_639_3) {
          return true
        }
        if (locale === c.iso_639_2) {
          return true
        }
        if (locale === c.iso_639_1) {
          return true
        }
        if (locale === c.title) {
          return true
        }
        return false
      })
      if (!loc) {
        windowPost.error("post-edit failed to find the locale: ", locale)
        return
      }
      const tid = t.id
      const tv = Object.values(store.translationValue).find(c => c.translation_id === tid && c.locale_id === loc.id)
      if (!tv) {
        windowPost.error("post-edit failed to find the translation-value ")
        return
      }
      api.translationValue.update(tv.id, { id: tv.id, value, ...context && { context_key: context } })
      break
    default:
      windowPost.error("unhandled kind for message:", e.data)
  }
}
window.addEventListener("message", handleMsg)





const mergeMap = <K extends DBKeyValue, V extends DB[K]>(key: K, value: V) => {
  if (!key) {
    console.error('key is required in mergeField')
    return
  }
  if (!value) {
    console.error('value is required in mergeField')
    return
  }
  db.update((s) => {
    return {
      ...s,
      [key]: {
        ...s[key],
        ...value,
      },
    }
  })
}

// Keys in db that are of type Record<string, T>
type DBKeyValue = keyof Omit<DB, 'serverInfo' | 'responseStates' | 'login'>

const replaceField = <K extends DBKeyValue, V extends DB[K]['s']>(
  key: K,
  value: V,
  id: string,
  kind: 'get' | 'update' | 'create' | 'soft-delete'
) => {
  if (!key) {
    console.error('key is required in replaceField')
    return
  }
  if (!value) {
    console.error('value is required in replaceField')
    return
  }
  if (!id) {
    console.error('id is required in replaceField')
    return
  }
  db.update((s) => {
    if (kind === 'create' || kind === 'update' || kind === 'soft-delete') {

      switch (key) {
        // FIXME: fix the typing above, so that typescript infers the correct value here.
        // Or is this perhaps fixed in ts4.6, which may be better at these things.
        case 'translationValue':
          if ('locale_id' in value) {
            windowPost.translationMessage('update/create' as any, value, s)
          }
      }
    }
    return {
      ...s,
      [key]: {
        ...s[key],
        [id]: value,
      },
    }
  })
}

/* 
  Returns typed functions for:
  - Create
  - Get
  - List

  yes, that is not really all of the cruds...
*/
function CrudFactory<Payload extends {}, K extends DBKeyValue, UpdatePayload = Payload>(
  storeKey: K,
  subPath?: string
) {
  return {
    get: apiGetFactory(subPath || storeKey, storeKey),
    list: apiGetListFactory(subPath || storeKey, storeKey),
    create: apiCreateFactory<Payload, K>(subPath || storeKey, storeKey),
    update: apiUpdateFactory<UpdatePayload, K>(subPath || storeKey, storeKey),
    delete: apiDeleteFactory<K>(subPath || storeKey, storeKey),
  }
}

function apiGetListFactory<K extends DBKeyValue>(subPath: string, storeKey: K) {
  return async (options?: ApiFetchOptions) => {
    db.update((s) => {
      return {
        ...s,
        responseStates: {
          ...s.responseStates,
          [storeKey]: {
            ...s.responseStates?.[storeKey as any],
            loading: true,
          },
        },
      }
    })
    const res = await fetchApi<DB[K]>(
      subPath,
      (e) => mergeMap(storeKey, e),
      options
    )
    checkError(res[1])
    db.update((s) => {
      return {
        ...s,
        ...(!res[1] &&
          !!res[0].data && {
          [storeKey]: { ...s[storeKey], ...res[0].data },
        }),
        responseStates: {
          ...s.responseStates,
          [storeKey]: {
            ...s.responseStates?.[storeKey as any],
            loading: false,
            error: res[1],
          },
        },
      }
    })
    return res
  }
}
const checkError = (apiError?: ApiDef.APIError | null) => {
  if (!apiError) {
    return apiError
  }
  toastApiErr(apiError)
  if (apiError.error?.code?.includes('Authentication required')) {
    db.update((s) => {
      const login = { ...s.login, ok: false }
      localStorage.setItem('login-response', JSON.stringify(login))
      return {
        ...s,
        login,
      }
    })
  }
  return apiError
}

function apiGetFactory<K extends DBKeyValue>(subPath: string, storeKey: K) {
  return (id: string, options?: ApiFetchOptions) =>
    fetchApi<DB[K]>(
      subPath + id,
      (e: any) => replaceField(storeKey, e, e.id, 'get'),
      options
    ).then((r) => {
      checkError(r[1])
      return r
    })
}
function apiCreateFactory<Payload extends {}, K extends DBKeyValue>(
  subPath: string,
  storeKey: K
) {
  return async (body: Payload, options?: ApiFetchOptions) => {
    db.update((s) => ({
      ...s,
      responseStates: { ...s.responseStates, [storeKey]: { loading: true } },
    }))
    const [res, err] = await fetchApi<DB[K]['s']>(
      subPath,
      (e) => e.id && replaceField(storeKey, e, e.id, 'create'),
      {
        method: methods.POST,
        body,
        ...options,
      }
    )
    checkError(err)
    if (!err && res) {
      toast({ kind: 'info', title: 'Success', message: `${storeKey} created` })
    }
    db.update((s) => ({
      ...s,
      responseStates: {
        ...s.responseStates,
        [storeKey]: { loading: false, error: err },
      },
    }))

    return [res, err]
  }
}

function apiUpdateFactory<Payload extends {}, K extends DBKeyValue>(
  subPath: string,
  storeKey: K
) {
  if (!subPath) {
    subPath = storeKey
  }
  return (id: string, body: Payload, options?: ApiFetchOptions) => {
    db.update((s) => ({
      ...s,
      responseStates: { ...s.responseStates, [storeKey]: { loading: true } }
    }))
    if (!(body as any).id) {
      console.warn("No id set on body for request", subPath, id, body);
      (body as any).id = id
    }

    return fetchApi<DB[K]['s']>(
      subPath,// + '/' + id,
      (e) => e.id && replaceField(storeKey, e, e.id, 'update'),
      {
        method: methods.PUT,
        body,
        ...options,
      }
    ).then((r) => {
      const [d, err] = r
      checkError(err)

      if (d?.data?.id) {
        replaceField(storeKey, r[0].data, d.data.id, 'update')
      }
      if (!err && d) {
        toast({ kind: 'info', title: 'Success', message: `${storeKey} updated` })
      }
      db.update((s) => ({
        ...s,
        responseStates: {
          ...s.responseStates,
          [storeKey]: { loading: false, error: err },
        },
      }))
      return r
    })
  }
}

function apiDeleteFactory<K extends DBKeyValue>(subPath: string, storeKey: K) {
  if (!subPath) {
    subPath = storeKey
  }
  return (id: string, body: ApiDef.DeleteInput, options?: ApiFetchOptions) =>
    fetchApi<DB[K]['s']>(
      subPath + '/' + id,
      (e) => e.id && replaceField(storeKey, e, e.id, 'soft-delete'),
      {
        method: methods.DELETE,
        body,
        ...options,
      }
    ).then((r) => {
      checkError(r[1])
      db.update((s) => ({
        ...s,
        responseStates: {
          ...s.responseStates,
          [storeKey]: { loading: false, error: r[1] },
        },
      }))
      return r
    })
}
