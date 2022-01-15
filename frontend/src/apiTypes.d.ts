declare namespace ApiDef {
    export interface APIError {
        code?: ErrorCodes;
        details?: {
            [key: string]: any;
        };
        error?: string;
    }
    export interface Category {
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
        description?: string;
        /**
         * Unique identifier of the entity
         */
        id: string;
        key?: string;
        project_id?: string;
        title?: string;
        translation_ids?: string[];
        /**
         * Time of which the entity was updated, if any
         */
        updatedAt?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updatedBy?: string;
    }
    export interface CategoryInput {
        description?: string;
        key: string; // ^[^\s]*$
        project_id: string;
        title: string;
    }
    export type CreatorSource = string;
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
    export type ErrorCodes = string;
    export interface ImportInput {
        [name: string]: any;
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
    export interface MissingTranslation {
        /**
         * The reported category (may not exist), as reported by the client.
         */
        category?: string;
        category_id?: string;
        /**
         * Number of times it has been reported.
         */
        count?: number; // int64
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
        first_user_agent?: string;
        /**
         * Unique identifier of the entity
         */
        id: string;
        latest_user_agent?: string;
        /**
         * The reported locale (may not exist), as reported by the client.
         */
        locale?: string;
        locale_id?: string;
        /**
         * The reported project (may not exist), as reported by the client.
         */
        project?: string;
        project_id?: string;
        /**
         * The reported translation (may not exist), as reported by the client.
         */
        translation?: string;
        translation_id?: string;
        /**
         * Time of which the entity was updated, if any
         */
        updatedAt?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updatedBy?: string;
    }
    export interface Project {
        category_ids?: string[];
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
        description?: string;
        /**
         * Unique identifier of the entity
         */
        id: string;
        included_tags?: string[];
        short_name?: string;
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
    export interface ProjectInput {
        description?: string;
        short_name: string;
        title: string;
    }
    export interface ReportMissingInput {
        [name: string]: string;
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
        category?: string;
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
        description?: string;
        /**
         * Unique identifier of the entity
         */
        id: string;
        key?: string;
        parent_translation?: string;
        tags?: string[];
        title?: string;
        /**
         * Time of which the entity was updated, if any
         */
        updatedAt?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updatedBy?: string;
        value_ids?: string[];
        variables?: {
            [name: string]: {
                [key: string]: any;
            };
        };
    }
    export interface TranslationInput {
        category_id: string;
        description?: string;
        key: string; // ^[^\s]*$
        title?: string;
        /**
         * key/value type. The value can be any type, but the key must a string.
         */
        variables?: {
            [name: string]: {
                [key: string]: any;
            };
        };
    }
    export interface TranslationValue {
        context?: {
            [name: string]: string;
        };
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
         * locale ID
         */
        locale_id?: string;
        source?: CreatorSource;
        /**
         * Translation ID
         */
        translation_id?: string;
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
    }
    export interface TranslationValueInput {
        context?: {
            [name: string]: string;
        };
        locale_id: string;
        translation_id: string;
        value: string;
    }
    export interface UpdateTranslationValueInput {
        value: string;
    }
}
declare namespace ApiPaths {
    namespace CreateTranslation {
        export interface BodyParameters {
            TranslationInput: Parameters.TranslationInput;
        }
        namespace Parameters {
            export type TranslationInput = ApiDef.TranslationInput;
        }
    }
    namespace GetExport {
        namespace Parameters {
            /**
             * Used to set the export-format.
             * ### `i18n`
             * The output is formatted to be i18n-compliant.
             * ### `raw`
             * The output is not converted, and all data is outputted.
             * The short-alias for this parameter is: `f`
             *
             */
            export type Format = "raw" | "i18n";
            /**
             * The parameter can be any of the Locale's ID, iso639_1, iso639_2, iso639_3, or ietf_tag.
             * By leaving out the parameters, all locales will be returned.
             * The short-alias for this parameter is: `l`
             * **Future: By setting `locale_id=auto`, the server will infer the locale from the browsers headers.**
             *
             */
            export type Locale = string;
            /**
             * Used to set which key in output for the locale that should be used.
             * The parameter can be any of the Locale's ID, iso639_1, iso639_2, iso639_3, or ietf_tag.
             * The short-alias for this parameter is: `p`
             *
             */
            export type LocaleKey = string;
            /**
             * Disables flattening of the outputet map
             *
             */
            export type NoFlatten = boolean;
            /**
             * The parameter can be any of the Project's ID or ShortName.
             * The short-alias for this parameter is: `p`
             *
             */
            export type Project = string;
        }
        export interface QueryParameters {
            project?: /**
             * The parameter can be any of the Project's ID or ShortName.
             * The short-alias for this parameter is: `p`
             *
             */
            Parameters.Project;
            format?: /**
             * Used to set the export-format.
             * ### `i18n`
             * The output is formatted to be i18n-compliant.
             * ### `raw`
             * The output is not converted, and all data is outputted.
             * The short-alias for this parameter is: `f`
             *
             */
            Parameters.Format;
            no_flatten?: /**
             * Disables flattening of the outputet map
             *
             */
            Parameters.NoFlatten;
            locale_key?: /**
             * Used to set which key in output for the locale that should be used.
             * The parameter can be any of the Locale's ID, iso639_1, iso639_2, iso639_3, or ietf_tag.
             * The short-alias for this parameter is: `p`
             *
             */
            Parameters.LocaleKey;
            locale?: /**
             * The parameter can be any of the Locale's ID, iso639_1, iso639_2, iso639_3, or ietf_tag.
             * By leaving out the parameters, all locales will be returned.
             * The short-alias for this parameter is: `l`
             * **Future: By setting `locale_id=auto`, the server will infer the locale from the browsers headers.**
             *
             */
            Parameters.Locale;
        }
        namespace Responses {
            export interface $200 {
            }
        }
    }
    namespace ImportTranslations {
        export interface BodyParameters {
            ImportInput?: Parameters.ImportInput;
        }
        namespace Parameters {
            /**
             * If set, a dry-run will occur, and the result is returned.
             *
             */
            export type Dry = boolean;
            export type ImportInput = ApiDef.ImportInput;
            /**
             * The format of the imported object.
             * If set to auto, the server will attempt to find the format for you.
             *
             */
            export type Kind = "i18n" | "auto";
            /**
             * The parameter can be any of the Locale's ID, iso639_1, iso639_2, iso639_3, or ietf_tag.
             *
             */
            export type Locale = string;
            /**
             * The parameter can be any of the Project's ID or ShortName.
             *
             */
            export type Project = string;
        }
        export interface PathParameters {
            kind: /**
             * The format of the imported object.
             * If set to auto, the server will attempt to find the format for you.
             *
             */
            Parameters.Kind;
            project: /**
             * The parameter can be any of the Project's ID or ShortName.
             *
             */
            Parameters.Project;
            locale: /**
             * The parameter can be any of the Locale's ID, iso639_1, iso639_2, iso639_3, or ietf_tag.
             *
             */
            Parameters.Locale;
        }
        export interface QueryParameters {
            dry?: /**
             * If set, a dry-run will occur, and the result is returned.
             *
             */
            Parameters.Dry;
        }
        namespace Responses {
            export interface $200 {
            }
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
    namespace ReportMissing {
        export interface BodyParameters {
            ReportMissingInput?: Parameters.ReportMissingInput;
        }
        namespace Parameters {
            /**
             * The parameter can be any of the Locale's ID, iso639_1, iso639_2, iso639_3, or ietf_tag.
             *
             */
            export type Locale = string;
            /**
             * The parameter can be any of the Project's ID or ShortName.
             *
             */
            export type Project = string;
            export type ReportMissingInput = ApiDef.ReportMissingInput;
        }
        export interface PathParameters {
            project: /**
             * The parameter can be any of the Project's ID or ShortName.
             *
             */
            Parameters.Project;
            locale: /**
             * The parameter can be any of the Locale's ID, iso639_1, iso639_2, iso639_3, or ietf_tag.
             *
             */
            Parameters.Locale;
        }
        namespace Responses {
            export interface $200 {
            }
        }
    }
    namespace UpdateTranslationValue {
        export interface BodyParameters {
            UpdateTranslationValueInput: Parameters.UpdateTranslationValueInput;
        }
        namespace Parameters {
            export type UpdateTranslationValueInput = ApiDef.UpdateTranslationValueInput;
        }
    }
}
declare namespace ApiResponses {
    export type ApiError = ApiDef.APIError;
    export type CategoriesResponse = ApiDef.Category[];
    export type CategoryResponse = ApiDef.Category;
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
    export type TranslationValueResponse = ApiDef.TranslationValue;
    export type TranslationValuesResponse = ApiDef.TranslationValue[];
    export type TranslationsResponse = ApiDef.Translation[];
}
