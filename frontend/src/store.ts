import { writable } from 'svelte/store'
import type { Updater } from 'svelte/store'
// TODO: try to remove these lodash-functions
import merge from 'merge'
import debounce from './util/debounce'
import deepEqual from './util/isEqual'

const parseStrOrObject = <T extends {}>(s: string | T | null | undefined) => {
  if (!s) {
    return [null, null] as const
  }
  if (typeof s !== 'string') {
    return [s as T, null] as const
  }
  try {
    const j = JSON.parse(s) as T
    return [j, null] as const
  } catch (err) {
    return [null, `failed to parse string into object '${s}': ${err}'`] as const
  }
}

export type StoreState<V = null, VK extends string = string> = {
  __didChange?: boolean
  __validationMessage?: Partial<Record<VK, string>>
  __validationPayload?: V
}

export type Store<T extends {}, V = null, VK extends string = string> = T &
  StoreState

/* 
TODO: use a form-library instead. 
  This turned out too complex. 
  Keep this store simple, and only use it only for the api, and perhaps some state.
 */
function createStore<T extends {}, V = null, VK extends string = string>({
  storage: _storage,
  initialValue,
  validator,
}: {
  storage?: AppStorage<T> | {
    key: string,
    ignoreKeys?: string[]
  }
  validator?: (
    t: Store<T, V>
  ) => [null, null] | [V, null] | [null, Partial<Record<VK, string>>]
  initialValue?: T
} = {}) {
  type S = Store<T, V, VK>
  let fromStorageValue: T | null = null
  let restoreValue = initialValue

  const storage: AppStorage<T> | null = _storage?.key
    ? {
      getItem: (key) => localStorage.getItem(key),
      // TODO: throttle saving
      setItem: (k, v) => localStorage.setItem(k, JSON.stringify(v)),
      ..._storage,
    }
    : null

  if (storage) {
    const str = storage.getItem(storage.key)
    const [parsed, err] = parseStrOrObject<T>(str)
    if (err) {
      console.error(err)
    } else if (parsed) {
      fromStorageValue = initialValue ? merge({}, initialValue, parsed) : parsed
    }
  }
  const validate = (value: S): S => {
    if (!value) {
      return value
    }
    if (!validator) {
      return value
    }
    const [v, errMsg] = validator(value)
    return {
      ...value,
      __validationMessage: errMsg,
      __validationPayload: v,
    } as any
  }
  const {
    update: _update,
    subscribe,
    set: _set,
  } = writable<S>(fromStorageValue ?? (initialValue as any))
  const _saveToStorageNow = (value: T) => {
    if (!storage || !_storage?.key) {
      return
    }
    if (storage.ignoreKeys?.length) {
      const v: any = Object.keys(value).reduce(
        (r, key) => {
          if (storage.ignoreKeys?.includes(key)) {
            return r
          }
          r[key] = value[key]

          return r
        }, {}
      )
      storage?.setItem(_storage.key, v)
      return
    }
    storage?.setItem(_storage.key, value)
  }
  const saveToStorage =
    !!storage &&
    debounce(_saveToStorageNow, 2000, {
      leading: true,
    })

  function didChange(existing) {
    const {
      __didChange: _,
      __validationMessage: _2,
      __validationPayload: _3,
      storeState,
      ...restNs
    } = existing
    const changed = !deepEqual(restNs, restoreValue)
    return changed
  }

  const update = (updater: Updater<S>, storeState?: StoreState) => {
    _update((s) => {
      let ns = storeState ? { ...updater(s), storeState } : updater(s)

      if (ns === s) {
        return ns
      }
      // If there is no validator, we assume we dont care about changes.
      if (!storeState && validator) {
        if (restoreValue) {
          ns.__didChange = didChange(ns)
        }
      }
      ns = validate(ns)

      if (saveToStorage) {
        storeState ? _saveToStorageNow(ns) : saveToStorage(ns)
      }
      return ns
    })
  }

  /** Like update, but also resets all store-state  */
  const restore = (state?: S) => {
    const s = update(
      () => {
        if (!state) {
          return merge({}, restoreValue)
        }
        const ns = merge({}, initialValue, state)
        restoreValue = ns
        return ns
      },
      {
        __didChange: false,
        __validationMessage: undefined,
        __validationPayload: undefined,
      }
    )
    return s
  }
  const set = (s: S) => {
    s.__didChange = didChange(s)
    if (saveToStorage) {
      saveToStorage(s)
    }
    s = validate(s)
    _set(merge({}, s))
  }
  const reset = () => {
    const s = validate({
      __didChange: false,
      __validationMessage: undefined,
      __validationPayload: undefined,
      ...(initialValue as any),
      // ...(initialValue as any),
    })

    _set(s)
    if (storage && _storage) {
      _saveToStorageNow(s)
    }
  }

  return {
    restore,
    reset,
    subscribe,
    update,
    set,
  }
}

export interface AppStorage<T extends {}> {
  getItem: (key: string) => string | T | null
  setItem: (key: string, value: T) => void
  key: string
  ignoreKeys?: string[]
}

export default createStore
