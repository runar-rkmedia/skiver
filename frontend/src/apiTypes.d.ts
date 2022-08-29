declare namespace ApiDef {
    export interface APIError {
        details?: {
            [key: string]: any;
        };
        error?: Error;
    }
    export interface Category {
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
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
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
    }
    export interface CategoryInput {
        description?: string;
        key: string; // ^[^\s]*$
        project_id: string;
        title: string;
    }
    export interface CategoryTreeNode {
        categories?: {
            [name: string]: CategoryTreeNode;
        };
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
        /**
         * If set, the item is considered deleted. The item will normally not get deleted from the database,
         * but it may if cleanup is required.
         */
        deleted?: string; // date-time
        description?: string;
        /**
         * TODO: change to map
         */
        exists?: boolean;
        /**
         * Unique identifier of the entity
         */
        id: string;
        key?: string;
        project_id?: string;
        title?: string;
        translation_ids?: string[];
        translations?: {
            [name: string]: ExtendedTranslation;
        };
        /**
         * Time of which the entity was updated, if any
         */
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
    }
    /**
     * Change stores information about a changed item
     */
    export interface Change {
        from?: {
            [key: string]: any;
        };
        path?: string[];
        to?: {
            [key: string]: any;
        };
        type?: string;
    }
    export interface ChangePasswordInput {
        new_password?: string;
        password: string;
    }
    /**
     * Changelog stores a list of changed items
     */
    export type Changelog = /* Change stores information about a changed item */ Change[];
    export interface CreateSnapshotInput {
        description?: string;
        project_id: string;
        tag: string; // ^[a-zA-Z0-9-_.]{3,36}$
    }
    export interface CreateTokenInput {
        description: string;
        /**
         * Duration in hours of which the token should be valid
         */
        ttl_hours: number;
    }
    export type CreatorSource = string;
    export interface DeleteInput {
        /**
         * Time of which the item at the earliest can be permanently deleted.
         *
         */
        expiryDate?: string; // date-time
        /**
         * If set, will bring the item back from the deletion-queue.
         */
        undelete?: boolean;
    }
    export interface DiffSnapshotInput {
        a: SnapshotSelector;
        b: SnapshotSelector;
        format?: "raw" | "i18n" | "typescript";
    }
    export interface Entity {
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
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
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
    }
    export interface Error {
        code?: ErrorCodes;
        error?: string;
    }
    export type ErrorCodes = string;
    export interface ExtendedCategory {
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
        /**
         * If set, the item is considered deleted. The item will normally not get deleted from the database,
         * but it may if cleanup is required.
         */
        deleted?: string; // date-time
        description?: string;
        /**
         * TODO: change to map
         */
        exists?: boolean;
        /**
         * Unique identifier of the entity
         */
        id: string;
        key?: string;
        project_id?: string;
        title?: string;
        translation_ids?: string[];
        translations?: {
            [name: string]: ExtendedTranslation;
        };
        /**
         * Time of which the entity was updated, if any
         */
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
    }
    export interface ExtendedProject {
        categories?: {
            [name: string]: ExtendedCategory;
        };
        category_ids?: string[];
        category_tree?: CategoryTreeNode;
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
        /**
         * If set, the item is considered deleted. The item will normally not get deleted from the database,
         * but it may if cleanup is required.
         */
        deleted?: string; // date-time
        description?: string;
        exists?: boolean;
        /**
         * Unique identifier of the entity
         */
        id: string;
        included_tags?: string[];
        locales?: {
            [name: string]: /**
             * # See https://en.wikipedia.org/wiki/Language_code for more information
             * TODO: consider supporting other standards here, like Windows(?), which seem to have their own thing.
             */
            Locale;
        };
        short_name?: string;
        snapshots?: {
            [name: string]: ProjectSnapshotMeta;
        };
        title?: string;
        /**
         * Time of which the entity was updated, if any
         */
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
    }
    export interface ExtendedTranslation {
        aliases?: string[];
        category?: string;
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
        /**
         * If set, the item is considered deleted. The item will normally not get deleted from the database,
         * but it may if cleanup is required.
         */
        deleted?: string; // date-time
        description?: string;
        exists?: boolean;
        /**
         * Unique identifier of the entity
         */
        id: string;
        key?: string;
        parent_translation?: string;
        references?: string[];
        tags?: string[];
        title?: string;
        /**
         * Time of which the entity was updated, if any
         */
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
        value_ids?: string[];
        values?: {
            [name: string]: TranslationValue;
        };
        variables?: {
            [name: string]: {
                [key: string]: any;
            };
        };
    }
    export interface ImportInput {
        [name: string]: any;
    }
    export interface JoinInput {
        password: string;
        /**
         * example:
         * abc123
         */
        username: string; // ^[^\s]*$
    }
    /**
     * # See https://en.wikipedia.org/wiki/Language_code for more information
     * TODO: consider supporting other standards here, like Windows(?), which seem to have their own thing.
     */
    export interface Locale {
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
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
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
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
    export interface LocaleSetting {
        /**
         * If set, will allow registered translation-services to translate from other languages to this locale.
         * This might help speed up translations for new locales.
         * See the Config or Organization-settings for instructions on how to set up translation-services.
         *
         * Organization-settings are not yet available.
         *
         * TODO: implement organization-settings
         */
        auto_translation?: boolean;
        /**
         * If set, the locale will be visible for editing.
         */
        enabled?: boolean;
        /**
         * If set, the associated translations will be published in releases.
         * This is useful for when adding new locales, and one don't want to publish it to users until it is complete
         */
        publish?: boolean;
    }
    export interface LocaleSettingInput {
        /**
         * If set, will allow registered translation-services to translate from other languages to this locale.
         * This might help speed up translations for new locales.
         * See the Config or Organization-settings for instructions on how to set up translation-services.
         *
         * Organization-settings are not yet available.
         *
         * TODO: implement organization-settings
         */
        auto_translation?: boolean;
        /**
         * If set, the locale will be visible for editing.
         */
        enabled?: boolean;
        /**
         * If set, the associated translations will be published in releases.
         * This is useful for when adding new locales, and one don't want to publish it to users until it is complete
         */
        publish?: boolean;
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
        active?: boolean;
        can_create_locales?: boolean;
        can_create_organization?: boolean;
        can_create_projects?: boolean;
        can_create_translations?: boolean;
        can_create_users?: boolean;
        can_manage_snapshots?: boolean;
        can_update_locales?: boolean;
        can_update_organization?: boolean;
        can_update_projects?: boolean;
        can_update_translations?: boolean;
        can_update_users?: boolean;
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
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
        organization?: Organization;
        /**
         * If set, the user must change the password before the account can be used
         */
        temporary_password?: boolean;
        /**
         * Time of which the entity was updated, if any
         */
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
        username?: string;
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
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
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
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
    }
    export interface OkResponse {
        ok: boolean;
    }
    export interface Organization {
        created_at?: string; // date-time
        created_by?: string;
        deleted?: string; // date-time
        description?: string;
        id?: string;
        /**
         * This will allow anybody with the id to create a standard user, and join the organization
         * The first user to join, gets priviliges to administer the organization.
         */
        join_id?: string;
        join_id_expires?: string; // date-time
        title?: string;
        updated_at?: string; // date-time
        updated_by?: string;
    }
    export interface OrganizationInput {
        title: string;
    }
    export interface Project {
        category_ids?: string[];
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
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
        locales?: {
            [name: string]: LocaleSetting;
        };
        short_name?: string;
        snapshots?: {
            [name: string]: ProjectSnapshotMeta;
        };
        title?: string;
        /**
         * Time of which the entity was updated, if any
         */
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
    }
    export interface ProjectDiffResponse {
        a?: ProjectStats;
        b?: ProjectStats;
        diff?: /* Changelog stores a list of changed items */ Changelog;
    }
    export interface ProjectInput {
        description?: string;
        locales?: {
            [name: string]: LocaleSetting;
        };
        short_name: string; // ^[a-z1-9]*$
        title: string;
    }
    export interface ProjectSnapshot {
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
        /**
         * If set, the item is considered deleted. The item will normally not get deleted from the database,
         * but it may if cleanup is required.
         */
        deleted?: string; // date-time
        /**
         * Unique identifier of the entity
         */
        id: string;
        project?: ExtendedProject;
        project_hash?: number; // uint64
        /**
         * Time of which the entity was updated, if any
         */
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
    }
    export interface ProjectSnapshotMeta {
        created_at?: string; // date-time
        created_by?: string;
        description?: string;
        hash?: number; // uint64
        id?: string;
        uploadMeta?: UploadMeta[];
    }
    export interface ProjectStats {
        hash?: string;
        identi_hash?: number /* uint8 */[];
        project_id?: string;
        size?: number; // uint64
        size_humanized?: string;
        tag?: string;
    }
    export interface ReleaseInfo {
        assets_url?: string;
        body?: string;
        created_at?: string;
        draft?: boolean;
        html_url?: string;
        name?: string;
        prerelease?: boolean;
        published_at?: string;
        tag_name?: string;
        target_commitish?: string;
        upload_url?: string;
        url?: string;
    }
    export interface ReportMissingInput {
        [name: string]: string;
    }
    export interface ServerInfo {
        /**
         * Date of build
         */
        build_date?: string; // date-time
        /**
         * Size of database.
         */
        database_size?: number; // int64
        database_size_str?: string;
        /**
         * Short githash for current commit
         */
        git_hash?: string;
        /**
         * Hash of the current host. Should be semi-stable
         */
        host_hash?: string;
        /**
         * Server-instance. This will change on every restart.
         */
        instance?: string;
        latest_cli_release?: ReleaseInfo;
        latest_release?: ReleaseInfo;
        /**
         * The minimum version of skiver-cli that can be used with this server.
         * The is [semver](https://semver.org/)-compatible, but has a leading `v`, like `v1.2.3`
         */
        min_cli_version?: string;
        /**
         * When the server was started
         */
        server_started_at?: string; // date-time
        /**
         * Version-number for commit
         */
        version?: string;
    }
    export type SimpleUser = string;
    export interface SnapshotSelector {
        project_id: string;
        tag?: string;
    }
    export interface TokenResponse {
        /**
         * Description of user-generated-token, or for login-tokens, this will be the last User-Agent used
         */
        description?: string;
        expires?: string; // date-time
        issued?: string; // date-time
        token?: string;
    }
    export interface Translation {
        aliases?: string[];
        category?: string;
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
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
        references?: string[];
        tags?: string[];
        title?: string;
        /**
         * Time of which the entity was updated, if any
         */
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
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
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
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
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
        /**
         * The pre-interpolated value to use  with translations
         * example:
         * The {{productName}} fires up to {{count}} bullets of {{subject}}.
         */
        value?: string;
    }
    export interface TranslationValueInput {
        /**
         * If set, it will add/update the context for that key instead of the original value
         */
        context_key?: string; // ^[^\s]*$
        locale_id: string;
        translation_id: string;
        value?: string;
    }
    export interface UpdateCategoryInput {
        description?: string;
        id?: string;
        key?: string; // ^[^\s]*$
        project_id?: string;
        title?: string;
    }
    export interface UpdateOrganizationInput {
        id: string;
        join_id?: string;
        join_id_expires?: string; // date-time
    }
    export interface UpdateProjectInput {
        description?: string;
        id: string;
        locales?: {
            [name: string]: LocaleSettingInput;
        };
        short_name?: string; // ^[a-z1-9]*$
        title?: string;
    }
    export interface UpdateTranslationInput {
        description?: string;
        id: string;
        key?: string; // ^[^\s]*$
        title?: string;
        variables?: {
            [name: string]: any;
        };
    }
    export interface UpdateTranslationValueInput {
        /**
         * If set, it will add/update the context for that key instead of the original value
         */
        context_key?: string; // ^[^\s]*$
        id: string;
        value?: string;
    }
    export interface UploadMeta {
        id?: string;
        locale?: string;
        locale_key?: string;
        parent?: string;
        provider_id?: string;
        provider_name?: string;
        size?: number; // int64
        tag?: string;
        url?: string;
    }
    export interface User {
        /**
         * If not active, the account cannot be used until any issues are resolved.
         */
        active?: boolean;
        can_create_locales?: boolean;
        can_create_organization?: boolean;
        can_create_projects?: boolean;
        can_create_translations?: boolean;
        can_create_users?: boolean;
        can_manage_snapshots?: boolean;
        can_update_locales?: boolean;
        can_update_organization?: boolean;
        can_update_projects?: boolean;
        can_update_translations?: boolean;
        can_update_users?: boolean;
        /**
         * Time of which the entity was created in the database
         */
        created_at: string; // date-time
        /**
         * User id refering to the user who created the item
         */
        created_by?: string;
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
         * If set, the user must change the password before the account can be used
         */
        temporary_password?: boolean;
        /**
         * Time of which the entity was updated, if any
         */
        updated_at?: string; // date-time
        /**
         * User id refering to who created the item
         */
        updated_by?: string;
        username?: string;
    }
}
declare namespace ApiPaths {
    namespace ChangePassword {
        export interface BodyParameters {
            changePassword: Parameters.ChangePassword;
        }
        namespace Parameters {
            export type ChangePassword = ApiDef.ChangePasswordInput;
        }
    }
    namespace CreateOrganization {
        export interface BodyParameters {
            OrganizationInput: Parameters.OrganizationInput;
        }
        namespace Parameters {
            export type OrganizationInput = ApiDef.OrganizationInput;
        }
    }
    namespace CreateSnapshot {
        export interface BodyParameters {
            SnapshotInput?: Parameters.SnapshotInput;
        }
        namespace Parameters {
            export type SnapshotInput = ApiDef.CreateSnapshotInput;
        }
        namespace Responses {
            export type $200 = ApiResponses.SnapshotResponse;
        }
    }
    namespace CreateToken {
        export interface BodyParameters {
            CreateToken: Parameters.CreateToken;
        }
        namespace Parameters {
            export type CreateToken = ApiDef.CreateTokenInput;
        }
    }
    namespace CreateTranslation {
        export interface BodyParameters {
            TranslationInput: Parameters.TranslationInput;
        }
        namespace Parameters {
            export type TranslationInput = ApiDef.TranslationInput;
        }
    }
    namespace DeleteTranslation {
        export interface BodyParameters {
            DeleteInput?: Parameters.DeleteInput;
        }
        namespace Parameters {
            export type DeleteInput = ApiDef.DeleteInput;
            export type Id = string;
        }
        export interface PathParameters {
            id: Parameters.Id;
        }
    }
    namespace DiffSnapshots {
        export interface BodyParameters {
            SnapshotInput?: Parameters.SnapshotInput;
        }
        namespace Parameters {
            export type SnapshotInput = ApiDef.DiffSnapshotInput;
        }
        namespace Responses {
            export type $200 = ApiResponses.DiffResponse;
        }
    }
    namespace GetExport {
        namespace Parameters {
            /**
             * Used to set the export-format.
             * The short-alias for this parameter is: `f`
             * ### `i18n`
             * The output is formatted to be i18n-compliant.
             * ### `raw`
             * The output is not converted, and all data is outputted.
             * ### `typescript`
             * Outputs a typescript-object-map of translation-keys for use with translation-libraries. Information is inclued in the TSDOC for each key.
             *
             */
            export type Format = "raw" | "typescript" | "i18n";
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
             * To be visible anonomously, the organization-id should be provided.
             * If you want to use the logged in users organization for this, specify `me` instead. This special key is set here to make it clear from the url  thaat this requires login.
             *
             */
            export type Organization = string;
            /**
             * The parameter can be any of the Project's ID or ShortName.
             *
             */
            export type Project = string;
        }
        export interface PathParameters {
            project: /**
             * The parameter can be any of the Project's ID or ShortName.
             *
             */
            Parameters.Project;
            organization: /**
             * To be visible anonomously, the organization-id should be provided.
             * If you want to use the logged in users organization for this, specify `me` instead. This special key is set here to make it clear from the url  thaat this requires login.
             *
             */
            Parameters.Organization;
        }
        export interface QueryParameters {
            format?: /**
             * Used to set the export-format.
             * The short-alias for this parameter is: `f`
             * ### `i18n`
             * The output is formatted to be i18n-compliant.
             * ### `raw`
             * The output is not converted, and all data is outputted.
             * ### `typescript`
             * Outputs a typescript-object-map of translation-keys for use with translation-libraries. Information is inclued in the TSDOC for each key.
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
    namespace GetOrgByJoinID {
        namespace Parameters {
            /**
             * The join-id.
             *
             */
            export type Id = string;
        }
        export interface PathParameters {
            id: /**
             * The join-id.
             *
             */
            Parameters.Id;
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
            export type Kind = "i18n" | "describe" | "auto";
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
    namespace JoinOrganiztaion {
        export interface BodyParameters {
            JoinInput: Parameters.JoinInput;
        }
        namespace Parameters {
            /**
             * The join-id.
             *
             */
            export type Id = string;
            export type JoinInput = ApiDef.JoinInput;
        }
        export interface PathParameters {
            id: /**
             * The join-id.
             *
             */
            Parameters.Id;
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
    namespace UpdateOrganization {
        export interface BodyParameters {
            OrganizationUpdateInput: Parameters.OrganizationUpdateInput;
        }
        namespace Parameters {
            export type OrganizationUpdateInput = ApiDef.UpdateOrganizationInput;
        }
    }
    namespace UpdateProject {
        export interface BodyParameters {
            UpdateProject: Parameters.UpdateProject;
        }
        namespace Parameters {
            export type UpdateProject = ApiDef.UpdateProjectInput;
        }
    }
    namespace UpdateTranslation {
        export interface BodyParameters {
            TranslationUpdateInput: Parameters.TranslationUpdateInput;
        }
        namespace Parameters {
            export type TranslationUpdateInput = ApiDef.UpdateTranslationInput;
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
    export type DiffResponse = ApiDef.ProjectDiffResponse;
    export type JoinResponse = ApiDef.LoginResponse;
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
    export type OkResponse = ApiDef.OkResponse;
    export type OrganizationResponse = ApiDef.Organization;
    export type OrganizationsResponse = ApiDef.Organization[];
    export type ProjectResponse = ApiDef.Project;
    export type ProjectsResponse = ApiDef.Project[];
    export type ServerInfo = ApiDef.ServerInfo[];
    export interface SimpleUsersResponse {
        [name: string]: ApiDef.SimpleUser;
    }
    export type SnapshotResponse = ApiDef.ProjectSnapshot;
    export type SnapshotsResponse = ApiDef.ProjectSnapshot[];
    export type TokenResponse = ApiDef.TokenResponse;
    export type TranslationResponse = ApiDef.Translation;
    export type TranslationValueResponse = ApiDef.TranslationValue;
    export type TranslationValuesResponse = ApiDef.TranslationValue[];
    export type TranslationsResponse = ApiDef.Translation[];
    export interface UsersResponse {
        [name: string]: ApiDef.User;
    }
}
