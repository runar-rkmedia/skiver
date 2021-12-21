import { ApiFetchOptions, fetchApi, methods, wsSubscribe } from './apiFetcher'
import createStore from './store'
import { objectKeys } from 'simplytyped'

/**
 * Typedefintions are created by running `yarn gen`.
 *
 * This will use the generated swagger defintions from Gobyoall-api. (which again are created from running go generate)
 */

export type DB = {
  project: Record<string, ApiDef.Project>
  locale: Record<string, ApiDef.LoginResponse>
  login: ApiDef.LoginResponse
  serverInfo: ApiDef.ServerInfo
  responseStates: Pick<Record<keyof DB, { loading: boolean, error?: ApiDef.ApiError }>, 'project' | 'locale' | 'login'>
}

export const api = {
  serverInfo: (options?: ApiFetchOptions) =>
    fetchApi<ApiDef.ServerInfo>(
      'serverInfo',
      (e) => db.update((s) => ({ ...s, serverInfo: e })),
      options
    ),
  project: CrudFactory<ApiDef.ProjectInput, 'project'>('project'),
  locale: CrudFactory<ApiDef.LocaleInput, 'locale'>('locale'),
  login: {
    get: () => fetchApi<ApiDef.LoginResponse>(
      'login',
      (login) => db.update(s => ({ ...s, login })),
      { method: "GET" }),
    post: async (args: ApiDef.LoginInput) => {
      db.update(s => (
        {
          ...s,
          responseStates: {
            ...s.responseStates,
            login: {
              ...s.responseStates.login,
              loading: true
            }
          }
        }))
      const [res, err] = await fetchApi<ApiDef.LoginResponse>(
        'login',
        (login) => db.update(s => ({ ...s, login })),
        { method: "POST", body: args })
      db.update(s => (
        {
          ...s,
          responseStates: {
            ...s.responseStates,
            login: {
              loading: false,
              ...res,
              ...!!err && { error: err }
            }
          }
        }))
      return [res, err]
    },
  }
} as const

export const db = createStore<DB, null>({
  initialValue: objectKeys(api).reduce((r, k) => {
    r[k] = {}
    return r
  }, {
    responseStates: objectKeys(api).reduce((r, k) => ({ ...r, [k]: { loading: false } }), {})
  } as DB),
})


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
    db.update(s => {
      return {
        ...s,
        responseStates: {
          ...s.responseStates,
          [storeKey]: {
            ...s.responseStates?.[storeKey],
            loading: true,
          }
        }
      }
    })
    const res = await fetchApi<DB[K]>(
      subPath,
      (e) => mergeMap(storeKey, e),
      options
    )
    db.update(s => {
      return {
        ...s,
        ...(!res[1] && !!res[0].data && {

          [storeKey]: { ...s[storeKey], ...res[0].data },
        }),
        responseStates: {
          ...s.responseStates,
          [storeKey]: {
            ...s.responseStates?.[storeKey],
            loading: false,
            error: res[1],
          }
        }
      }
    })
    return res
  }
}

function apiGetFactory<K extends DBKeyValue>(subPath: string, storeKey: K) {
  return (id: string, options?: ApiFetchOptions) =>
    fetchApi<DB[K]>(
      subPath + id,
      (e: any) => replaceField(storeKey, e, e.id),
      options
    )
}
function apiCreateFactory<Payload extends {}, K extends DBKeyValue>(
  subPath: string,
  storeKey: K
) {
  return (body: Payload, options?: ApiFetchOptions) =>
    fetchApi<DB[K]['s']>(subPath, (e) => replaceField(storeKey, e, e.id), {
      method: methods.POST,
      body,
      ...options,
    })
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
    )
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
    )
}
