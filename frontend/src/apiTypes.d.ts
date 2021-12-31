declare namespace ApiDef {
  export interface APIError {
    code?: ErrorCodes
    details?: {
      [key: string]: any
    }
    error?: string
  }
  export interface Category {
    /**
     * Time of which the entity was created in the database
     */
    createdAt: string // date-time
    /**
     * User id refering to the user who created the item
     */
    createdBy?: string
    /**
     * If set, the item is considered deleted. The item will normally not get deleted from the database,
     * but it may if cleanup is required.
     */
    deleted?: string // date-time
    description?: string
    /**
     * Unique identifier of the entity
     */
    id: string
    key?: string
    project_id?: string
    title?: string
    /**
     * Time of which the entity was updated, if any
     */
    updatedAt?: string // date-time
    /**
     * User id refering to who created the item
     */
    updatedBy?: string
  }
  export interface CategoryInput {
    description?: string
    key: string
    project_id: string
    title: string
  }
  export type CreatorSource = string
  export interface Entity {
    /**
     * Time of which the entity was created in the database
     */
    createdAt: string // date-time
    /**
     * User id refering to the user who created the item
     */
    createdBy?: string
    /**
     * If set, the item is considered deleted. The item will normally not get deleted from the database,
     * but it may if cleanup is required.
     */
    deleted?: string // date-time
    /**
     * Unique identifier of the entity
     */
    id: string
    /**
     * Time of which the entity was updated, if any
     */
    updatedAt?: string // date-time
    /**
     * User id refering to who created the item
     */
    updatedBy?: string
  }
  export type ErrorCodes = string
  /**
   * # See https://en.wikipedia.org/wiki/Language_code for more information
   * TODO: consider supporting other standards here, like Windows(?), which seem to have their own thing.
   */
  export interface Locale {
    /**
     * Time of which the entity was created in the database
     */
    createdAt: string // date-time
    /**
     * User id refering to the user who created the item
     */
    createdBy?: string
    /**
     * If set, the item is considered deleted. The item will normally not get deleted from the database,
     * but it may if cleanup is required.
     */
    deleted?: string // date-time
    /**
     * List of other Locales in preferred order for fallbacks
     */
    fallbacks?: string[]
    /**
     * Unique identifier of the entity
     */
    id: string
    /**
     * Represents the IETF language tag, e.g. en / en-US
     */
    ietf?: string
    /**
     * Represents the ISO-639-1 string, e.g. en
     */
    iso_639_1?: string
    /**
     * Represents the ISO-639-2 string, e.g. eng
     */
    iso_639_2?: string
    /**
     * Represents the ISO-639-3 string, e.g. eng
     */
    iso_639_3?: string
    title?: string
    /**
     * Time of which the entity was updated, if any
     */
    updatedAt?: string // date-time
    /**
     * User id refering to who created the item
     */
    updatedBy?: string
  }
  export interface LocaleInput {
    /**
     * List of other Locales in preferred order for fallbacks
     */
    ietf_tag: string
    iso639_1: string
    iso639_2: string
    iso639_3: string
    title: string
  }
  export interface LoginInput {
    password: string
    /**
     * example:
     * abc123
     */
    username: string // ^[^\s]*$
  }
  export interface LoginResponse {
    /**
     * If not active, the account cannot be used until any issues are resolved.
     */
    Active?: boolean
    /**
     * Time of which the entity was created in the database
     */
    createdAt: string // date-time
    /**
     * User id refering to the user who created the item
     */
    createdBy?: string
    /**
     * If set, the item is considered deleted. The item will normally not get deleted from the database,
     * but it may if cleanup is required.
     */
    deleted?: string // date-time
    expires?: string // date-time
    expires_in?: string
    /**
     * Unique identifier of the entity
     */
    id: string
    ok?: boolean
    /**
     * If set, the user must change the password before the account can be used
     */
    temporary_password?: boolean
    /**
     * Time of which the entity was updated, if any
     */
    updatedAt?: string // date-time
    /**
     * User id refering to who created the item
     */
    updatedBy?: string
    userName?: string
  }
  export interface Project {
    category_ids?: string[]
    /**
     * Time of which the entity was created in the database
     */
    createdAt: string // date-time
    /**
     * User id refering to the user who created the item
     */
    createdBy?: string
    /**
     * If set, the item is considered deleted. The item will normally not get deleted from the database,
     * but it may if cleanup is required.
     */
    deleted?: string // date-time
    description?: string
    /**
     * Unique identifier of the entity
     */
    id: string
    included_tags?: string[]
    title?: string
    /**
     * Time of which the entity was updated, if any
     */
    updatedAt?: string // date-time
    /**
     * User id refering to who created the item
     */
    updatedBy?: string
  }
  export interface ProjectInput {
    description?: string
    title: string
  }
  export interface ServerInfo {
    /**
     * Date of build
     */
    BuildDate?: string // date-time
    /**
     * Size of database.
     */
    DatabaseSize?: number // int64
    DatabaseSizeStr?: string
    /**
     * Short githash for current commit
     */
    GitHash?: string
    /**
     * When the server was started
     */
    ServerStartedAt?: string // date-time
    /**
     * Version-number for commit
     */
    Version?: string
  }
  export interface Translation {
    aliases?: string[]
    category?: string
    context?: string
    /**
     * Time of which the entity was created in the database
     */
    createdAt: string // date-time
    /**
     * User id refering to the user who created the item
     */
    createdBy?: string
    /**
     * If set, the item is considered deleted. The item will normally not get deleted from the database,
     * but it may if cleanup is required.
     */
    deleted?: string // date-time
    description?: string
    /**
     * Unique identifier of the entity
     */
    id: string
    key?: string
    parent_translation?: string
    tags?: string[]
    title?: string
    /**
     * Time of which the entity was updated, if any
     */
    updatedAt?: string // date-time
    /**
     * User id refering to who created the item
     */
    updatedBy?: string
    variables?: {
      [name: string]: {
        [key: string]: any
      }
    }
  }
  export interface TranslationInput {
    category_id: string
    context?: string
    description?: string
    key: string
    title?: string
  }
  export interface TranslationValue {
    /**
     * Time of which the entity was created in the database
     */
    createdAt: string // date-time
    /**
     * User id refering to the user who created the item
     */
    createdBy?: string
    /**
     * If set, the item is considered deleted. The item will normally not get deleted from the database,
     * but it may if cleanup is required.
     */
    deleted?: string // date-time
    /**
     * Unique identifier of the entity
     */
    id: string
    /**
     * locale ID
     */
    locale_id?: string
    source?: CreatorSource
    /**
     * Translation ID
     */
    translation_id?: string
    /**
     * Time of which the entity was updated, if any
     */
    updatedAt?: string // date-time
    /**
     * User id refering to who created the item
     */
    updatedBy?: string
    /**
     * The pre-interpolated value to use  with translations
     * example:
     * The {{productName}} fires up to {{count}} bullets of {{subject}}.
     */
    value?: string
  }
  export interface TranslationValueInput {
    locale_id: string
    translation_id: string
    value: string
  }
}
declare namespace ApiPaths {
  namespace ListLocale {
    export interface BodyParameters {
      LocaleInput: Parameters.LocaleInput
    }
    namespace Parameters {
      export type LocaleInput = ApiDef.LocaleInput
    }
  }
  namespace Login {
    export interface BodyParameters {
      LoginInput: Parameters.LoginInput
    }
    namespace Parameters {
      export type LoginInput = ApiDef.LoginInput
    }
  }
}
declare namespace ApiResponses {
  export type ApiError = ApiDef.APIError
  export type CategoriesResponse = ApiDef.Category[]
  export type CategoryResponse = ApiDef.Category
  export type LocaleResponse
  /**
   * # See https://en.wikipedia.org/wiki/Language_code for more information
   * TODO: consider supporting other standards here, like Windows(?), which seem to have their own thing.
   */ = ApiDef.Locale
  export type LocalesResponse
  /**
   * # See https://en.wikipedia.org/wiki/Language_code for more information
   * TODO: consider supporting other standards here, like Windows(?), which seem to have their own thing.
   */ = ApiDef.Locale[]
  export type LoginResponse = ApiDef.LoginResponse
  export type ProjectResponse = ApiDef.Project
  export type ProjectsResponse = ApiDef.Project[]
  export type ServerInfo = ApiDef.ServerInfo[]
  export type TranslationResponse = ApiDef.Translation
  export type TranslationValueResponse = ApiDef.TranslationValue
  export type TranslationValuesResponse = ApiDef.TranslationValue[]
  export type TranslationsResponse = ApiDef.Translation[]
}
