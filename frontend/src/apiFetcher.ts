import { apiUrl, appUrl } from "util/appConstants"

export const contentTypes = {
  json: 'application/json',
  toml: 'application/toml',
  yaml: 'text/vnd.yaml',
} as const

export const methods = {
  POST: 'POST',
  GET: 'GET',
  PUT: 'PUT',
  DELETE: 'DELETE',
} as const


export const wsUrl = appUrl('/ws/').replace('http', 'ws')
export type ApiFetchOptions = {
  /** if set, will run the updater-function even on errors */
  runUpdaterOnErr?: boolean
  /** Body, as json. Can either be stringified or an object, in which case it will be stringified */
  body?: {} | string
  /** HTTP-method */
  method?: string
  /** Allows to use JMES-path on the result. This will be handled by the api.
   *
   * NOTE: if set, the updater-function will not run by default.
   */
  jmespath?: string
}

export async function fetchApi<T extends {}>(
  subPath: string,
  updater: (data: T) => void,
  {
    method = methods.GET,
    body,
    jmespath,
    runUpdaterOnErr,
  }: ApiFetchOptions = {}
) {
  const sub = subPath.replace(/^\/?/, '/').replace(/\/?$/, '/')
  const opts: RequestInit = {
    method,
    headers: {
      accept: contentTypes.json,
      'content-type': contentTypes.json,
      ...(!!jmespath && {
        'jmes-path': jmespath,
      }),
    },
    ...(!!body && {
      body: typeof body === 'string' ? body : JSON.stringify(body),
    }),
  }
  const url = apiUrl(sub)
  const result: {
    data: T
  } = {} as any
  let response: Response | null = null
  try {
    response = await fetch(url, opts)
    const contentType = response.headers.get('content-type') || ''
    if (contentType.includes(contentTypes.json)) {
      const JSON = await response.json()
      if (JSON && !jmespath) {
        if (response.status < 400 || runUpdaterOnErr) updater(JSON)
      }
      if (response.status >= 400) {
        return [null, JSON as ApiResponses.ApiError] as const
      }
      result.data = JSON
    } else {

      const err: ApiResponses.ApiError = {
        error: { error: "Unexpected result from api: " + response.status }, details: {
          statusCode: response.status,
        }
      }
      return [null, err] as const
    }
  } catch (err) {
    console.error(`fetchApi error for ${subPath}: ${err.message}`, {
      subPath,
      url,
      opts,
      err,
      response,
    })
    return [
      result,
      {
        error: err.message as string,
        originalError: err,
        code: response?.status || 'NoStatusReceived',
      } as ApiResponses.ApiError & { originalError: Error },
    ] as const
  }
  return [result, null] as const
}

export function serializeDate(date: Date) {
  return date.toISOString()
}
export function deserializeDate(dateStr: string) {
  return new Date(dateStr)
}

let wsDisconnects = 0
let wsFails = 0
export const wsSubscribe = (options: {
  onMessage: (msg: WsMessage) => void
  onClose?: () => void
  autoReconnect: boolean
}) => {
  const { onMessage, onClose, autoReconnect } = options
  if (!window['WebSocket']) {
    console.error('Your browser does not support WebSocket')
    return
  }
  try {
    const conn = new WebSocket(wsUrl)
    conn.onerror = function(evt) {
      console.error('[ws] connection error: ', evt)
    }
    conn.onclose = function(evt) {
      console.debug('[ws]: connection closed', evt)
      onClose?.()
      wsDisconnects++
      if (autoReconnect) {
        setTimeout(() => wsSubscribe(options), 1000 * (wsDisconnects + wsFails))
      }
    }
    conn.onmessage = function(evt) {
      try {
        const json = JSON.parse(evt.data)
        onMessage(json)
      } catch (err) {
        console.error('Failed to parse json-message\n', err)
      }
    }
  } catch (err) {
    console.error('Failed in wsSubscribe ', err)
    wsFails++
  }
}

type Ws<K extends string, V extends string, T> = {
  kind: K
  variant: V
  contents: T
}

type verbs = 'create' | 'update' | 'soft-delete'

type WsProject = Ws<'project', verbs, ApiDef.Project>

type WsLocale = Ws<'locale', verbs, ApiDef.Locale>

type WsTranslationValue = Ws<'translationValue', verbs, ApiDef.TranslationValue>
type WsTranslation = Ws<'translation', verbs, ApiDef.Translation>

type WsMessage =
  | Ws<'dist', 'chache', null> // Only used in development
  | WsProject
  | WsLocale
  | WsTranslation
  | WsTranslationValue
