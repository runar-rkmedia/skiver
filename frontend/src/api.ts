import { fetchApi, methods, wsSubscribe } from './apiFetcher'
import type { ApiFetchOptions } from './apiFetcher'
import createStore from './store'
import type { AnyFunc } from 'simplytyped'
import { parseDate } from 'dates'
import { isPast } from 'date-fns'
import { derived } from 'svelte/store'
export function objectKeys<T extends object>(obj: T) {
  return Object.keys(obj) as Array<keyof T>
}

/**
 * Typedefintions are created by running `yarn gen`.
 *
 * This will use the generated swagger defintions from Gobyoall-api. (which again are created from running go generate)
 */

export type DB = {
  project: Record<string, ApiDef.Project>
  category: Record<string, ApiDef.Category>
  translationValue: Record<string, ApiDef.TranslationValue>
  locale: Record<string, ApiDef.Locale>
  translation: Record<string, ApiDef.Translation>
  login: ApiDef.LoginResponse
  serverInfo: ApiDef.ServerInfo
  responseStates: Pick<
    Record<keyof DB, { loading: boolean; error?: ApiDef.APIError }>,
    'project' | 'locale' | 'login'
  >
}


export const api = {
  serverInfo: (options?: ApiFetchOptions) =>
    fetchApi<ApiDef.ServerInfo>(
      'serverInfo',
      (e) => db.update((s) => ({ ...s, serverInfo: e })),
      options
    ),
  translation: CrudFactory<ApiDef.TranslationInput, 'translation'>('translation'),
  project: CrudFactory<ApiDef.ProjectInput, 'project'>('project'),
  category: CrudFactory<ApiDef.CategoryInput, 'category'>('category'),
  translationValue: CrudFactory<ApiDef.TranslationValueInput, 'translationValue'>('translationValue'),
  locale: CrudFactory<ApiDef.LocaleInput, 'locale'>('locale'),
  login: {
    get: () =>
      fetchApi<ApiDef.LoginResponse>(
        'login',
        (login) => {
          localStorage.setItem('login-response', JSON.stringify(login))
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
  const r = {
    ok: false,
    userName: '',
    createdAt: '',
    updatedAt: '',
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

type ExtendedTranslationValue = ApiDef.TranslationValue
type ExtendedTranslation = ApiDef.Translation & { values: Record<string, ExtendedTranslationValue> }
type ExtendedCategory = ApiDef.Category & { translations: Record<string, ExtendedTranslation> }
type ExtendedProject = ApiDef.Project & { categories: Record<string, ExtendedCategory> }

export const projects = derived(db, ($db) => Object.values($db.project).reduce(
  (r, project: ExtendedProject) => {
    project.categories = Object.values($db.category).reduce(
      (rc, c: ExtendedCategory) => {
        if (c.project_id !== project.id) {
          return rc
        }
        c.translations = Object.values($db.translation).reduce(
          (rt, t: ExtendedTranslation) => {
            if (t.category !== c.id) {
              return rt
            }
            t.values = Object.values($db.translationValue).reduce(
              (rtv, tv) => {
                if (tv.translation_id !== t.id) {
                  return rtv
                }
                // NOTE: translations are indexed by their locale-id, not their id.
                rtv[tv.locale_id!] = tv
                return rtv
              }, {}
            )
            rt[t.id] = t
            return rt
          }, {}
        )
        rc[c.id] = c

        return rc
      }, {}
    )
    r[project.id] = project
    return r
  }, {} as Record<string, ExtendedProject>
))

wsSubscribe({
  onMessage: (msg) => {
    if (!msg.contents) {
      return
    }
    if (typeof msg.contents !== 'object') {
      return
    }

    if (msg.contents.id) {
      replaceField(msg.kind, msg.contents as any, msg.contents.id)
    }
  },
  autoReconnect: true,
})

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
  id: string
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
function CrudFactory<Payload extends {}, K extends DBKeyValue>(
  storeKey: K,
  subPath?: string
) {
  return {
    get: apiGetFactory(subPath || storeKey, storeKey),
    list: apiGetListFactory(subPath || storeKey, storeKey),
    create: apiCreateFactory<Payload, K>(subPath || storeKey, storeKey),
    update: apiUpdateFactory<Payload, K>(subPath || storeKey, storeKey),
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
  console.log('checkErr')
  if (!apiError) {
    return apiError
  }
  if (apiError.code?.includes("Authentication required")) {
    db.update(s => {
      const login = { ...s.login, ok: false }
      localStorage.setItem('login-response', JSON.stringify(login))
      return {
        ...s, login
      }
    })
  }
  return apiError
}

function apiGetFactory<K extends DBKeyValue>(subPath: string, storeKey: K) {
  return (id: string, options?: ApiFetchOptions) =>
    fetchApi<DB[K]>(
      subPath + id,
      (e: any) => replaceField(storeKey, e, e.id),
      options
    ).then(r => { checkError(r[1]); return r })
}
function apiCreateFactory<Payload extends {}, K extends DBKeyValue>(
  subPath: string,
  storeKey: K
) {
  return async (body: Payload, options?: ApiFetchOptions) => {

    db.update(s => ({ ...s, responseStates: { ...s.responseStates, [storeKey]: { loading: true } } }))
    const result = await fetchApi<DB[K]['s']>(subPath, (e) => replaceField(storeKey, e, e.id), {
      method: methods.POST,
      body,
      ...options,
    })
    checkError(result[1])
    db.update(s => ({ ...s, responseStates: { ...s.responseStates, [storeKey]: { loading: false, error: result[1] } } }))

    return result
  }
}

function apiUpdateFactory<Payload extends {}, K extends DBKeyValue>(
  subPath: string,
  storeKey: K
) {
  if (!subPath) {
    subPath = storeKey
  }
  return (id: string, body: Payload, options?: ApiFetchOptions) =>
    fetchApi<DB[K]['s']>(
      subPath + '/' + id,
      (e) => replaceField(storeKey, e, e.id),
      {
        method: methods.PUT,
        body,
        ...options,
      }
    ).then(r => { checkError(r[1]); return r })
}

function apiDeleteFactory<K extends DBKeyValue>(subPath: string, storeKey: K) {
  if (!subPath) {
    subPath = storeKey
  }
  return (id: string, options?: ApiFetchOptions) =>
    fetchApi<DB[K]['s']>(
      subPath + '/' + id,
      (e) => replaceField(storeKey, e, e.id),
      {
        method: methods.DELETE,
        ...options,
      }
    ).then(r => { checkError(r[1]); return r })
}
