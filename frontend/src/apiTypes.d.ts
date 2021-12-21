declare namespace ApiDef {
    export interface ApiError {
        code?: string;
        details?: {
            [key: string]: any;
        };
        error?: string;
    }
    export interface Entity {
        /**
         * Time of which the entity was created in the database
         */
        createdAt: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        createdBy?: string;
        /**
         * If set, the item is considered deleted. The item will normally not get deleted from the database,
         * but it may if cleanup is required.
         */
        deleted?: string; // date-time
        /**
         * Unique identifier of the entity
         */
        id: string;
        /**
         * Time of which the entity was updated, if any
         */
        updatedAt?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updatedBy?: string;
    }
    /**
     * # See https://en.wikipedia.org/wiki/Language_code for more information
     * TODO: consider supporting other standards here, like Windows(?), which seem to have their own thing.
     */
    export interface Locale {
        /**
         * Time of which the entity was created in the database
         */
        createdAt: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        createdBy?: string;
        /**
         * If set, the item is considered deleted. The item will normally not get deleted from the database,
         * but it may if cleanup is required.
         */
        deleted?: string; // date-time
        /**
         * List of other Locales in preferred order for fallbacks
         */
        fallbacks?: string[];
        /**
         * Unique identifier of the entity
         */
        id: string;
        /**
         * Represents the IETF language tag, e.g. en / en-US
         */
        ietf?: string;
        /**
         * Represents the ISO-639-1 string, e.g. en
         */
        iso_639_1?: string;
        /**
         * Represents the ISO-639-2 string, e.g. eng
         */
        iso_639_2?: string;
        /**
         * Represents the ISO-639-3 string, e.g. eng
         */
        iso_639_3?: string;
        title?: string;
        /**
         * Time of which the entity was updated, if any
         */
        updatedAt?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updatedBy?: string;
    }
    export interface LocaleInput {
        /**
         * List of other Locales in preferred order for fallbacks
         */
        ietf_tag: string;
        iso639_1: string;
        iso639_2: string;
        iso639_3: string;
        title: string;
    }
    export interface LoginInput {
        password: string;
        /**
         * example:
         * abc123
         */
        username: string; // ^[^\s]*$
    }
    export interface LoginResponse {
        /**
         * If not active, the account cannot be used until any issues are resolved.
         */
        Active?: boolean;
        /**
         * Time of which the entity was created in the database
         */
        createdAt: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        createdBy?: string;
        /**
         * If set, the item is considered deleted. The item will normally not get deleted from the database,
         * but it may if cleanup is required.
         */
        deleted?: string; // date-time
        expires?: string; // date-time
        expires_in?: string;
        /**
         * Unique identifier of the entity
         */
        id: string;
        ok?: boolean;
        /**
         * If set, the user must change the password before the account can be used
         */
        temporary_password?: boolean;
        /**
         * Time of which the entity was updated, if any
         */
        updatedAt?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updatedBy?: string;
        userName?: string;
    }
    export interface Project {
        /**
         * Time of which the entity was created in the database
         */
        createdAt: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        createdBy?: string;
        /**
         * If set, the item is considered deleted. The item will normally not get deleted from the database,
         * but it may if cleanup is required.
         */
        deleted?: string; // date-time
        /**
         * example:
         * Project-description
         */
        description?: string;
        /**
         * Unique identifier of the entity
         */
        id: string;
        /**
         * If present, any translations with tags matching will also be included in the exported translations
         * If the project contains conflicting translations, the project has presedence.
         * example:
         * [
         *   "actions",
         *   "general"
         * ]
         */
        included_tags?: string[];
        /**
         * example:
         * My Great Project
         */
        title: string;
        /**
         * Time of which the entity was updated, if any
         */
        updatedAt?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updatedBy?: string;
    }
    export interface ProjectInput {
        /**
         * example:
         * Project-description
         */
        description?: string;
        /**
         * If present, any translations with tags matching will also be included in the exported translations
         * If the project contains conflicting translations, the project has presedence.
         * example:
         * [
         *   "actions",
         *   "general"
         * ]
         */
        included_tags?: string[];
        /**
         * example:
         * My Great Project
         */
        title: string;
    }
    export interface ServerInfo {
        /**
         * Date of build
         */
        BuildDate?: string; // date-time
        /**
         * Size of database.
         */
        DatabaseSize?: number; // int64
        DatabaseSizeStr?: string;
        /**
         * Short githash for current commit
         */
        GitHash?: string;
        /**
         * When the server was started
         */
        ServerStartedAt?: string; // date-time
        /**
         * Version-number for commit
         */
        Version?: string;
    }
    export interface Translation {
        aliases?: string[];
        /**
         * Used as a variation for the key
         */
        context?: string;
        /**
         * Time of which the entity was created in the database
         */
        createdAt: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        createdBy?: string;
        /**
         * If set, the item is considered deleted. The item will normally not get deleted from the database,
         * but it may if cleanup is required.
         */
        deleted?: string; // date-time
        /**
         * Description for the key, its use and where the key is used.
         */
        description?: string;
        /**
         * Unique identifier of the entity
         */
        id: string;
        /**
         * Final part of the identifiying key.
         * With the example-input, the complete generated key would be store.product.description
         * example:
         * description
         */
        key?: string;
        locale_id?: string;
        /**
         * Can be a dot-separated path-like string
         * example:
         * store.products
         */
        prefix?: string;
        project?: string;
        tags?: string[];
        /**
         * Title with short description of the key
         */
        title?: string;
        /**
         * Time of which the entity was updated, if any
         */
        updatedAt?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updatedBy?: string;
        /**
         * The pre-interpolated value to use  with translations
         * example:
         * The {{productName}} fires up to {{count}} bullets of {{subject}}.
         */
        value?: string;
        /**
         * Variables used within the translation.
         * This helps with giving translators more context,
         * The value for the translation will be used in examples.
         * example:
         * {
         *   "count": 3,
         *   "productName": "X-Buster",
         *   "subject": "compressed solar energy"
         * }
         */
        variables?: {
            [name: string]: {
                [key: string]: any;
            };
        };
    }
    export interface TranslationInput {
        aliases?: string[];
        /**
         * Used as a variation for the key
         */
        context?: string;
        /**
         * Description for the key, its use and where the key is used.
         */
        description?: string;
        /**
         * Final part of the identifiying key.
         * With the example-input, the complete generated key would be store.product.description
         * example:
         * description
         */
        key?: string;
        locale_id?: string;
        /**
         * Can be a dot-separated path-like string
         * example:
         * store.products
         */
        prefix?: string;
        project?: string;
        tags?: string[];
        /**
         * Title with short description of the key
         */
        title?: string;
        /**
         * The pre-interpolated value to use  with translations
         * example:
         * The {{productName}} fires up to {{count}} bullets of {{subject}}.
         */
        value?: string;
        /**
         * Variables used within the translation.
         * This helps with giving translators more context,
         * The value for the translation will be used in examples.
         * example:
         * {
         *   "count": 3,
         *   "productName": "X-Buster",
         *   "subject": "compressed solar energy"
         * }
         */
        variables?: {
            [name: string]: {
                [key: string]: any;
            };
        };
    }
}
declare namespace ApiPaths {
    namespace CreateProject {
        export interface BodyParameters {
            Body: Parameters.Body;
        }
        namespace Parameters {
            export type Body = ApiDef.ProjectInput;
        }
    }
    namespace CreateTranslation {
        export interface BodyParameters {
            Body: Parameters.Body;
        }
        namespace Parameters {
            export type Body = ApiDef.TranslationInput;
        }
    }
    namespace ListLocale {
        export interface BodyParameters {
            LocaleInput: Parameters.LocaleInput;
        }
        namespace Parameters {
            export type LocaleInput = ApiDef.LocaleInput;
        }
    }
    namespace Login {
        export interface BodyParameters {
            LoginInput: Parameters.LoginInput;
        }
        namespace Parameters {
            export type LoginInput = ApiDef.LoginInput;
        }
    }
}
declare namespace ApiResponses {
    export type ApiError = ApiDef.ApiError;
    export type LocaleResponse = /**
     * # See https://en.wikipedia.org/wiki/Language_code for more information
     * TODO: consider supporting other standards here, like Windows(?), which seem to have their own thing.
     */
    ApiDef.Locale;
    export type LocalesResponse = /**
     * # See https://en.wikipedia.org/wiki/Language_code for more information
     * TODO: consider supporting other standards here, like Windows(?), which seem to have their own thing.
     */
    ApiDef.Locale[];
    export type LoginResponse = ApiDef.LoginResponse;
    export type ProjectResponse = ApiDef.Project;
    export type ProjectsResponse = ApiDef.Project[];
    export type ServerInfo = ApiDef.ServerInfo[];
    export type TranslationResponse = ApiDef.Translation;
    export type TranslationsResponse = ApiDef.Translation[];
}
