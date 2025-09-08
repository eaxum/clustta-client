/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
import * as $protobuf from "protobufjs/minimal";

// Common aliases
const $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
const $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});

export const repository = $root.repository = (() => {

    /**
     * Namespace repository.
     * @exports repository
     * @namespace
     */
    const repository = {};

    repository.User = (function() {

        /**
         * Properties of a User.
         * @memberof repository
         * @interface IUser
         * @property {string|null} [id] User id
         * @property {number|Long|null} [mtime] User mtime
         * @property {string|null} [added_at] User added_at
         * @property {string|null} [username] User username
         * @property {string|null} [email] User email
         * @property {string|null} [first_name] User first_name
         * @property {string|null} [last_name] User last_name
         * @property {Uint8Array|null} [photo] User photo
         * @property {string|null} [role_id] User role_id
         * @property {boolean|null} [synced] User synced
         * @property {string|null} [role] User role
         */

        /**
         * Constructs a new User.
         * @memberof repository
         * @classdesc Represents a User.
         * @implements IUser
         * @constructor
         * @param {repository.IUser=} [properties] Properties to set
         */
        function User(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * User id.
         * @member {string} id
         * @memberof repository.User
         * @instance
         */
        User.prototype.id = "";

        /**
         * User mtime.
         * @member {number|Long} mtime
         * @memberof repository.User
         * @instance
         */
        User.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * User added_at.
         * @member {string} added_at
         * @memberof repository.User
         * @instance
         */
        User.prototype.added_at = "";

        /**
         * User username.
         * @member {string} username
         * @memberof repository.User
         * @instance
         */
        User.prototype.username = "";

        /**
         * User email.
         * @member {string} email
         * @memberof repository.User
         * @instance
         */
        User.prototype.email = "";

        /**
         * User first_name.
         * @member {string} first_name
         * @memberof repository.User
         * @instance
         */
        User.prototype.first_name = "";

        /**
         * User last_name.
         * @member {string} last_name
         * @memberof repository.User
         * @instance
         */
        User.prototype.last_name = "";

        /**
         * User photo.
         * @member {Uint8Array} photo
         * @memberof repository.User
         * @instance
         */
        User.prototype.photo = $util.newBuffer([]);

        /**
         * User role_id.
         * @member {string} role_id
         * @memberof repository.User
         * @instance
         */
        User.prototype.role_id = "";

        /**
         * User synced.
         * @member {boolean} synced
         * @memberof repository.User
         * @instance
         */
        User.prototype.synced = false;

        /**
         * User role.
         * @member {string} role
         * @memberof repository.User
         * @instance
         */
        User.prototype.role = "";

        /**
         * Creates a new User instance using the specified properties.
         * @function create
         * @memberof repository.User
         * @static
         * @param {repository.IUser=} [properties] Properties to set
         * @returns {repository.User} User instance
         */
        User.create = function create(properties) {
            return new User(properties);
        };

        /**
         * Encodes the specified User message. Does not implicitly {@link repository.User.verify|verify} messages.
         * @function encode
         * @memberof repository.User
         * @static
         * @param {repository.IUser} message User message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        User.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.added_at != null && Object.hasOwnProperty.call(message, "added_at"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.added_at);
            if (message.username != null && Object.hasOwnProperty.call(message, "username"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.username);
            if (message.email != null && Object.hasOwnProperty.call(message, "email"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.email);
            if (message.first_name != null && Object.hasOwnProperty.call(message, "first_name"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.first_name);
            if (message.last_name != null && Object.hasOwnProperty.call(message, "last_name"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.last_name);
            if (message.photo != null && Object.hasOwnProperty.call(message, "photo"))
                writer.uint32(/* id 8, wireType 2 =*/66).bytes(message.photo);
            if (message.role_id != null && Object.hasOwnProperty.call(message, "role_id"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.role_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 10, wireType 0 =*/80).bool(message.synced);
            if (message.role != null && Object.hasOwnProperty.call(message, "role"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.role);
            return writer;
        };

        /**
         * Encodes the specified User message, length delimited. Does not implicitly {@link repository.User.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.User
         * @static
         * @param {repository.IUser} message User message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        User.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a User message from the specified reader or buffer.
         * @function decode
         * @memberof repository.User
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.User} User
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        User.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.User();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.added_at = reader.string();
                        break;
                    }
                case 4: {
                        message.username = reader.string();
                        break;
                    }
                case 5: {
                        message.email = reader.string();
                        break;
                    }
                case 6: {
                        message.first_name = reader.string();
                        break;
                    }
                case 7: {
                        message.last_name = reader.string();
                        break;
                    }
                case 8: {
                        message.photo = reader.bytes();
                        break;
                    }
                case 9: {
                        message.role_id = reader.string();
                        break;
                    }
                case 10: {
                        message.synced = reader.bool();
                        break;
                    }
                case 11: {
                        message.role = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a User message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.User
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.User} User
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        User.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a User message.
         * @function verify
         * @memberof repository.User
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        User.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.added_at != null && message.hasOwnProperty("added_at"))
                if (!$util.isString(message.added_at))
                    return "added_at: string expected";
            if (message.username != null && message.hasOwnProperty("username"))
                if (!$util.isString(message.username))
                    return "username: string expected";
            if (message.email != null && message.hasOwnProperty("email"))
                if (!$util.isString(message.email))
                    return "email: string expected";
            if (message.first_name != null && message.hasOwnProperty("first_name"))
                if (!$util.isString(message.first_name))
                    return "first_name: string expected";
            if (message.last_name != null && message.hasOwnProperty("last_name"))
                if (!$util.isString(message.last_name))
                    return "last_name: string expected";
            if (message.photo != null && message.hasOwnProperty("photo"))
                if (!(message.photo && typeof message.photo.length === "number" || $util.isString(message.photo)))
                    return "photo: buffer expected";
            if (message.role_id != null && message.hasOwnProperty("role_id"))
                if (!$util.isString(message.role_id))
                    return "role_id: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            if (message.role != null && message.hasOwnProperty("role"))
                if (!$util.isString(message.role))
                    return "role: string expected";
            return null;
        };

        /**
         * Creates a User message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.User
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.User} User
         */
        User.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.User)
                return object;
            let message = new $root.repository.User();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.added_at != null)
                message.added_at = String(object.added_at);
            if (object.username != null)
                message.username = String(object.username);
            if (object.email != null)
                message.email = String(object.email);
            if (object.first_name != null)
                message.first_name = String(object.first_name);
            if (object.last_name != null)
                message.last_name = String(object.last_name);
            if (object.photo != null)
                if (typeof object.photo === "string")
                    $util.base64.decode(object.photo, message.photo = $util.newBuffer($util.base64.length(object.photo)), 0);
                else if (object.photo.length >= 0)
                    message.photo = object.photo;
            if (object.role_id != null)
                message.role_id = String(object.role_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            if (object.role != null)
                message.role = String(object.role);
            return message;
        };

        /**
         * Creates a plain object from a User message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.User
         * @static
         * @param {repository.User} message User
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        User.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.added_at = "";
                object.username = "";
                object.email = "";
                object.first_name = "";
                object.last_name = "";
                if (options.bytes === String)
                    object.photo = "";
                else {
                    object.photo = [];
                    if (options.bytes !== Array)
                        object.photo = $util.newBuffer(object.photo);
                }
                object.role_id = "";
                object.synced = false;
                object.role = "";
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.added_at != null && message.hasOwnProperty("added_at"))
                object.added_at = message.added_at;
            if (message.username != null && message.hasOwnProperty("username"))
                object.username = message.username;
            if (message.email != null && message.hasOwnProperty("email"))
                object.email = message.email;
            if (message.first_name != null && message.hasOwnProperty("first_name"))
                object.first_name = message.first_name;
            if (message.last_name != null && message.hasOwnProperty("last_name"))
                object.last_name = message.last_name;
            if (message.photo != null && message.hasOwnProperty("photo"))
                object.photo = options.bytes === String ? $util.base64.encode(message.photo, 0, message.photo.length) : options.bytes === Array ? Array.prototype.slice.call(message.photo) : message.photo;
            if (message.role_id != null && message.hasOwnProperty("role_id"))
                object.role_id = message.role_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            if (message.role != null && message.hasOwnProperty("role"))
                object.role = message.role;
            return object;
        };

        /**
         * Converts this User to JSON.
         * @function toJSON
         * @memberof repository.User
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        User.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for User
         * @function getTypeUrl
         * @memberof repository.User
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        User.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.User";
        };

        return User;
    })();

    repository.EntityType = (function() {

        /**
         * Properties of an EntityType.
         * @memberof repository
         * @interface IEntityType
         * @property {string|null} [id] EntityType id
         * @property {number|Long|null} [mtime] EntityType mtime
         * @property {string|null} [name] EntityType name
         * @property {string|null} [icon] EntityType icon
         * @property {boolean|null} [synced] EntityType synced
         */

        /**
         * Constructs a new EntityType.
         * @memberof repository
         * @classdesc Represents an EntityType.
         * @implements IEntityType
         * @constructor
         * @param {repository.IEntityType=} [properties] Properties to set
         */
        function EntityType(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * EntityType id.
         * @member {string} id
         * @memberof repository.EntityType
         * @instance
         */
        EntityType.prototype.id = "";

        /**
         * EntityType mtime.
         * @member {number|Long} mtime
         * @memberof repository.EntityType
         * @instance
         */
        EntityType.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * EntityType name.
         * @member {string} name
         * @memberof repository.EntityType
         * @instance
         */
        EntityType.prototype.name = "";

        /**
         * EntityType icon.
         * @member {string} icon
         * @memberof repository.EntityType
         * @instance
         */
        EntityType.prototype.icon = "";

        /**
         * EntityType synced.
         * @member {boolean} synced
         * @memberof repository.EntityType
         * @instance
         */
        EntityType.prototype.synced = false;

        /**
         * Creates a new EntityType instance using the specified properties.
         * @function create
         * @memberof repository.EntityType
         * @static
         * @param {repository.IEntityType=} [properties] Properties to set
         * @returns {repository.EntityType} EntityType instance
         */
        EntityType.create = function create(properties) {
            return new EntityType(properties);
        };

        /**
         * Encodes the specified EntityType message. Does not implicitly {@link repository.EntityType.verify|verify} messages.
         * @function encode
         * @memberof repository.EntityType
         * @static
         * @param {repository.IEntityType} message EntityType message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EntityType.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.icon != null && Object.hasOwnProperty.call(message, "icon"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.icon);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 5, wireType 0 =*/40).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified EntityType message, length delimited. Does not implicitly {@link repository.EntityType.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.EntityType
         * @static
         * @param {repository.IEntityType} message EntityType message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EntityType.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an EntityType message from the specified reader or buffer.
         * @function decode
         * @memberof repository.EntityType
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.EntityType} EntityType
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EntityType.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.EntityType();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.icon = reader.string();
                        break;
                    }
                case 5: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an EntityType message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.EntityType
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.EntityType} EntityType
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EntityType.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an EntityType message.
         * @function verify
         * @memberof repository.EntityType
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        EntityType.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.icon != null && message.hasOwnProperty("icon"))
                if (!$util.isString(message.icon))
                    return "icon: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates an EntityType message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.EntityType
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.EntityType} EntityType
         */
        EntityType.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.EntityType)
                return object;
            let message = new $root.repository.EntityType();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.icon != null)
                message.icon = String(object.icon);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from an EntityType message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.EntityType
         * @static
         * @param {repository.EntityType} message EntityType
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        EntityType.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.icon = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.icon != null && message.hasOwnProperty("icon"))
                object.icon = message.icon;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this EntityType to JSON.
         * @function toJSON
         * @memberof repository.EntityType
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        EntityType.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for EntityType
         * @function getTypeUrl
         * @memberof repository.EntityType
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        EntityType.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.EntityType";
        };

        return EntityType;
    })();

    repository.TaskType = (function() {

        /**
         * Properties of a TaskType.
         * @memberof repository
         * @interface ITaskType
         * @property {string|null} [id] TaskType id
         * @property {number|Long|null} [mtime] TaskType mtime
         * @property {string|null} [name] TaskType name
         * @property {string|null} [icon] TaskType icon
         * @property {boolean|null} [synced] TaskType synced
         */

        /**
         * Constructs a new TaskType.
         * @memberof repository
         * @classdesc Represents a TaskType.
         * @implements ITaskType
         * @constructor
         * @param {repository.ITaskType=} [properties] Properties to set
         */
        function TaskType(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * TaskType id.
         * @member {string} id
         * @memberof repository.TaskType
         * @instance
         */
        TaskType.prototype.id = "";

        /**
         * TaskType mtime.
         * @member {number|Long} mtime
         * @memberof repository.TaskType
         * @instance
         */
        TaskType.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TaskType name.
         * @member {string} name
         * @memberof repository.TaskType
         * @instance
         */
        TaskType.prototype.name = "";

        /**
         * TaskType icon.
         * @member {string} icon
         * @memberof repository.TaskType
         * @instance
         */
        TaskType.prototype.icon = "";

        /**
         * TaskType synced.
         * @member {boolean} synced
         * @memberof repository.TaskType
         * @instance
         */
        TaskType.prototype.synced = false;

        /**
         * Creates a new TaskType instance using the specified properties.
         * @function create
         * @memberof repository.TaskType
         * @static
         * @param {repository.ITaskType=} [properties] Properties to set
         * @returns {repository.TaskType} TaskType instance
         */
        TaskType.create = function create(properties) {
            return new TaskType(properties);
        };

        /**
         * Encodes the specified TaskType message. Does not implicitly {@link repository.TaskType.verify|verify} messages.
         * @function encode
         * @memberof repository.TaskType
         * @static
         * @param {repository.ITaskType} message TaskType message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TaskType.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.icon != null && Object.hasOwnProperty.call(message, "icon"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.icon);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 5, wireType 0 =*/40).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified TaskType message, length delimited. Does not implicitly {@link repository.TaskType.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.TaskType
         * @static
         * @param {repository.ITaskType} message TaskType message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TaskType.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a TaskType message from the specified reader or buffer.
         * @function decode
         * @memberof repository.TaskType
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.TaskType} TaskType
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TaskType.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.TaskType();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.icon = reader.string();
                        break;
                    }
                case 5: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a TaskType message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.TaskType
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.TaskType} TaskType
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TaskType.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a TaskType message.
         * @function verify
         * @memberof repository.TaskType
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        TaskType.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.icon != null && message.hasOwnProperty("icon"))
                if (!$util.isString(message.icon))
                    return "icon: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a TaskType message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.TaskType
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.TaskType} TaskType
         */
        TaskType.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.TaskType)
                return object;
            let message = new $root.repository.TaskType();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.icon != null)
                message.icon = String(object.icon);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a TaskType message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.TaskType
         * @static
         * @param {repository.TaskType} message TaskType
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        TaskType.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.icon = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.icon != null && message.hasOwnProperty("icon"))
                object.icon = message.icon;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this TaskType to JSON.
         * @function toJSON
         * @memberof repository.TaskType
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        TaskType.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for TaskType
         * @function getTypeUrl
         * @memberof repository.TaskType
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        TaskType.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.TaskType";
        };

        return TaskType;
    })();

    repository.Task = (function() {

        /**
         * Properties of a Task.
         * @memberof repository
         * @interface ITask
         * @property {string|null} [id] Task id
         * @property {number|Long|null} [mtime] Task mtime
         * @property {string|null} [created_at] Task created_at
         * @property {string|null} [name] Task name
         * @property {string|null} [description] Task description
         * @property {string|null} [extension] Task extension
         * @property {boolean|null} [is_resource] Task is_resource
         * @property {string|null} [status_id] Task status_id
         * @property {string|null} [task_type_id] Task task_type_id
         * @property {string|null} [entity_id] Task entity_id
         * @property {string|null} [assignee_id] Task assignee_id
         * @property {string|null} [assigner_id] Task assigner_id
         * @property {boolean|null} [is_link] Task is_link
         * @property {string|null} [pointer] Task pointer
         * @property {string|null} [preview_id] Task preview_id
         * @property {boolean|null} [trashed] Task trashed
         * @property {boolean|null} [synced] Task synced
         */

        /**
         * Constructs a new Task.
         * @memberof repository
         * @classdesc Represents a Task.
         * @implements ITask
         * @constructor
         * @param {repository.ITask=} [properties] Properties to set
         */
        function Task(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Task id.
         * @member {string} id
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.id = "";

        /**
         * Task mtime.
         * @member {number|Long} mtime
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Task created_at.
         * @member {string} created_at
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.created_at = "";

        /**
         * Task name.
         * @member {string} name
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.name = "";

        /**
         * Task description.
         * @member {string} description
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.description = "";

        /**
         * Task extension.
         * @member {string} extension
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.extension = "";

        /**
         * Task is_resource.
         * @member {boolean} is_resource
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.is_resource = false;

        /**
         * Task status_id.
         * @member {string} status_id
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.status_id = "";

        /**
         * Task task_type_id.
         * @member {string} task_type_id
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.task_type_id = "";

        /**
         * Task entity_id.
         * @member {string} entity_id
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.entity_id = "";

        /**
         * Task assignee_id.
         * @member {string} assignee_id
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.assignee_id = "";

        /**
         * Task assigner_id.
         * @member {string} assigner_id
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.assigner_id = "";

        /**
         * Task is_link.
         * @member {boolean} is_link
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.is_link = false;

        /**
         * Task pointer.
         * @member {string} pointer
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.pointer = "";

        /**
         * Task preview_id.
         * @member {string} preview_id
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.preview_id = "";

        /**
         * Task trashed.
         * @member {boolean} trashed
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.trashed = false;

        /**
         * Task synced.
         * @member {boolean} synced
         * @memberof repository.Task
         * @instance
         */
        Task.prototype.synced = false;

        /**
         * Creates a new Task instance using the specified properties.
         * @function create
         * @memberof repository.Task
         * @static
         * @param {repository.ITask=} [properties] Properties to set
         * @returns {repository.Task} Task instance
         */
        Task.create = function create(properties) {
            return new Task(properties);
        };

        /**
         * Encodes the specified Task message. Does not implicitly {@link repository.Task.verify|verify} messages.
         * @function encode
         * @memberof repository.Task
         * @static
         * @param {repository.ITask} message Task message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Task.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.created_at);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.name);
            if (message.description != null && Object.hasOwnProperty.call(message, "description"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.description);
            if (message.extension != null && Object.hasOwnProperty.call(message, "extension"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.extension);
            if (message.is_resource != null && Object.hasOwnProperty.call(message, "is_resource"))
                writer.uint32(/* id 7, wireType 0 =*/56).bool(message.is_resource);
            if (message.status_id != null && Object.hasOwnProperty.call(message, "status_id"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.status_id);
            if (message.task_type_id != null && Object.hasOwnProperty.call(message, "task_type_id"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.task_type_id);
            if (message.entity_id != null && Object.hasOwnProperty.call(message, "entity_id"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.entity_id);
            if (message.assignee_id != null && Object.hasOwnProperty.call(message, "assignee_id"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.assignee_id);
            if (message.assigner_id != null && Object.hasOwnProperty.call(message, "assigner_id"))
                writer.uint32(/* id 12, wireType 2 =*/98).string(message.assigner_id);
            if (message.is_link != null && Object.hasOwnProperty.call(message, "is_link"))
                writer.uint32(/* id 13, wireType 0 =*/104).bool(message.is_link);
            if (message.pointer != null && Object.hasOwnProperty.call(message, "pointer"))
                writer.uint32(/* id 14, wireType 2 =*/114).string(message.pointer);
            if (message.preview_id != null && Object.hasOwnProperty.call(message, "preview_id"))
                writer.uint32(/* id 15, wireType 2 =*/122).string(message.preview_id);
            if (message.trashed != null && Object.hasOwnProperty.call(message, "trashed"))
                writer.uint32(/* id 16, wireType 0 =*/128).bool(message.trashed);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 17, wireType 0 =*/136).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified Task message, length delimited. Does not implicitly {@link repository.Task.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Task
         * @static
         * @param {repository.ITask} message Task message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Task.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Task message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Task
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Task} Task
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Task.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Task();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.created_at = reader.string();
                        break;
                    }
                case 4: {
                        message.name = reader.string();
                        break;
                    }
                case 5: {
                        message.description = reader.string();
                        break;
                    }
                case 6: {
                        message.extension = reader.string();
                        break;
                    }
                case 7: {
                        message.is_resource = reader.bool();
                        break;
                    }
                case 8: {
                        message.status_id = reader.string();
                        break;
                    }
                case 9: {
                        message.task_type_id = reader.string();
                        break;
                    }
                case 10: {
                        message.entity_id = reader.string();
                        break;
                    }
                case 11: {
                        message.assignee_id = reader.string();
                        break;
                    }
                case 12: {
                        message.assigner_id = reader.string();
                        break;
                    }
                case 13: {
                        message.is_link = reader.bool();
                        break;
                    }
                case 14: {
                        message.pointer = reader.string();
                        break;
                    }
                case 15: {
                        message.preview_id = reader.string();
                        break;
                    }
                case 16: {
                        message.trashed = reader.bool();
                        break;
                    }
                case 17: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Task message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Task
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Task} Task
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Task.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Task message.
         * @function verify
         * @memberof repository.Task
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Task.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.created_at != null && message.hasOwnProperty("created_at"))
                if (!$util.isString(message.created_at))
                    return "created_at: string expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.description != null && message.hasOwnProperty("description"))
                if (!$util.isString(message.description))
                    return "description: string expected";
            if (message.extension != null && message.hasOwnProperty("extension"))
                if (!$util.isString(message.extension))
                    return "extension: string expected";
            if (message.is_resource != null && message.hasOwnProperty("is_resource"))
                if (typeof message.is_resource !== "boolean")
                    return "is_resource: boolean expected";
            if (message.status_id != null && message.hasOwnProperty("status_id"))
                if (!$util.isString(message.status_id))
                    return "status_id: string expected";
            if (message.task_type_id != null && message.hasOwnProperty("task_type_id"))
                if (!$util.isString(message.task_type_id))
                    return "task_type_id: string expected";
            if (message.entity_id != null && message.hasOwnProperty("entity_id"))
                if (!$util.isString(message.entity_id))
                    return "entity_id: string expected";
            if (message.assignee_id != null && message.hasOwnProperty("assignee_id"))
                if (!$util.isString(message.assignee_id))
                    return "assignee_id: string expected";
            if (message.assigner_id != null && message.hasOwnProperty("assigner_id"))
                if (!$util.isString(message.assigner_id))
                    return "assigner_id: string expected";
            if (message.is_link != null && message.hasOwnProperty("is_link"))
                if (typeof message.is_link !== "boolean")
                    return "is_link: boolean expected";
            if (message.pointer != null && message.hasOwnProperty("pointer"))
                if (!$util.isString(message.pointer))
                    return "pointer: string expected";
            if (message.preview_id != null && message.hasOwnProperty("preview_id"))
                if (!$util.isString(message.preview_id))
                    return "preview_id: string expected";
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                if (typeof message.trashed !== "boolean")
                    return "trashed: boolean expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a Task message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Task
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Task} Task
         */
        Task.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Task)
                return object;
            let message = new $root.repository.Task();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.created_at != null)
                message.created_at = String(object.created_at);
            if (object.name != null)
                message.name = String(object.name);
            if (object.description != null)
                message.description = String(object.description);
            if (object.extension != null)
                message.extension = String(object.extension);
            if (object.is_resource != null)
                message.is_resource = Boolean(object.is_resource);
            if (object.status_id != null)
                message.status_id = String(object.status_id);
            if (object.task_type_id != null)
                message.task_type_id = String(object.task_type_id);
            if (object.entity_id != null)
                message.entity_id = String(object.entity_id);
            if (object.assignee_id != null)
                message.assignee_id = String(object.assignee_id);
            if (object.assigner_id != null)
                message.assigner_id = String(object.assigner_id);
            if (object.is_link != null)
                message.is_link = Boolean(object.is_link);
            if (object.pointer != null)
                message.pointer = String(object.pointer);
            if (object.preview_id != null)
                message.preview_id = String(object.preview_id);
            if (object.trashed != null)
                message.trashed = Boolean(object.trashed);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a Task message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Task
         * @static
         * @param {repository.Task} message Task
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Task.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.created_at = "";
                object.name = "";
                object.description = "";
                object.extension = "";
                object.is_resource = false;
                object.status_id = "";
                object.task_type_id = "";
                object.entity_id = "";
                object.assignee_id = "";
                object.assigner_id = "";
                object.is_link = false;
                object.pointer = "";
                object.preview_id = "";
                object.trashed = false;
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.created_at != null && message.hasOwnProperty("created_at"))
                object.created_at = message.created_at;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.description != null && message.hasOwnProperty("description"))
                object.description = message.description;
            if (message.extension != null && message.hasOwnProperty("extension"))
                object.extension = message.extension;
            if (message.is_resource != null && message.hasOwnProperty("is_resource"))
                object.is_resource = message.is_resource;
            if (message.status_id != null && message.hasOwnProperty("status_id"))
                object.status_id = message.status_id;
            if (message.task_type_id != null && message.hasOwnProperty("task_type_id"))
                object.task_type_id = message.task_type_id;
            if (message.entity_id != null && message.hasOwnProperty("entity_id"))
                object.entity_id = message.entity_id;
            if (message.assignee_id != null && message.hasOwnProperty("assignee_id"))
                object.assignee_id = message.assignee_id;
            if (message.assigner_id != null && message.hasOwnProperty("assigner_id"))
                object.assigner_id = message.assigner_id;
            if (message.is_link != null && message.hasOwnProperty("is_link"))
                object.is_link = message.is_link;
            if (message.pointer != null && message.hasOwnProperty("pointer"))
                object.pointer = message.pointer;
            if (message.preview_id != null && message.hasOwnProperty("preview_id"))
                object.preview_id = message.preview_id;
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                object.trashed = message.trashed;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this Task to JSON.
         * @function toJSON
         * @memberof repository.Task
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Task.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Task
         * @function getTypeUrl
         * @memberof repository.Task
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Task.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Task";
        };

        return Task;
    })();

    repository.Entity = (function() {

        /**
         * Properties of an Entity.
         * @memberof repository
         * @interface IEntity
         * @property {string|null} [id] Entity id
         * @property {number|Long|null} [mtime] Entity mtime
         * @property {string|null} [created_at] Entity created_at
         * @property {string|null} [name] Entity name
         * @property {string|null} [entity_path] Entity entity_path
         * @property {string|null} [description] Entity description
         * @property {boolean|null} [trashed] Entity trashed
         * @property {string|null} [entity_type_id] Entity entity_type_id
         * @property {string|null} [parent_id] Entity parent_id
         * @property {string|null} [preview_id] Entity preview_id
         * @property {boolean|null} [synced] Entity synced
         * @property {boolean|null} [is_library] Entity is_library
         */

        /**
         * Constructs a new Entity.
         * @memberof repository
         * @classdesc Represents an Entity.
         * @implements IEntity
         * @constructor
         * @param {repository.IEntity=} [properties] Properties to set
         */
        function Entity(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Entity id.
         * @member {string} id
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.id = "";

        /**
         * Entity mtime.
         * @member {number|Long} mtime
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Entity created_at.
         * @member {string} created_at
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.created_at = "";

        /**
         * Entity name.
         * @member {string} name
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.name = "";

        /**
         * Entity entity_path.
         * @member {string} entity_path
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.entity_path = "";

        /**
         * Entity description.
         * @member {string} description
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.description = "";

        /**
         * Entity trashed.
         * @member {boolean} trashed
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.trashed = false;

        /**
         * Entity entity_type_id.
         * @member {string} entity_type_id
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.entity_type_id = "";

        /**
         * Entity parent_id.
         * @member {string} parent_id
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.parent_id = "";

        /**
         * Entity preview_id.
         * @member {string} preview_id
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.preview_id = "";

        /**
         * Entity synced.
         * @member {boolean} synced
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.synced = false;

        /**
         * Entity is_library.
         * @member {boolean} is_library
         * @memberof repository.Entity
         * @instance
         */
        Entity.prototype.is_library = false;

        /**
         * Creates a new Entity instance using the specified properties.
         * @function create
         * @memberof repository.Entity
         * @static
         * @param {repository.IEntity=} [properties] Properties to set
         * @returns {repository.Entity} Entity instance
         */
        Entity.create = function create(properties) {
            return new Entity(properties);
        };

        /**
         * Encodes the specified Entity message. Does not implicitly {@link repository.Entity.verify|verify} messages.
         * @function encode
         * @memberof repository.Entity
         * @static
         * @param {repository.IEntity} message Entity message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Entity.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.created_at);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.name);
            if (message.entity_path != null && Object.hasOwnProperty.call(message, "entity_path"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.entity_path);
            if (message.description != null && Object.hasOwnProperty.call(message, "description"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.description);
            if (message.trashed != null && Object.hasOwnProperty.call(message, "trashed"))
                writer.uint32(/* id 7, wireType 0 =*/56).bool(message.trashed);
            if (message.entity_type_id != null && Object.hasOwnProperty.call(message, "entity_type_id"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.entity_type_id);
            if (message.parent_id != null && Object.hasOwnProperty.call(message, "parent_id"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.parent_id);
            if (message.preview_id != null && Object.hasOwnProperty.call(message, "preview_id"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.preview_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 11, wireType 0 =*/88).bool(message.synced);
            if (message.is_library != null && Object.hasOwnProperty.call(message, "is_library"))
                writer.uint32(/* id 12, wireType 0 =*/96).bool(message.is_library);
            return writer;
        };

        /**
         * Encodes the specified Entity message, length delimited. Does not implicitly {@link repository.Entity.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Entity
         * @static
         * @param {repository.IEntity} message Entity message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Entity.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an Entity message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Entity
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Entity} Entity
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Entity.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Entity();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.created_at = reader.string();
                        break;
                    }
                case 4: {
                        message.name = reader.string();
                        break;
                    }
                case 5: {
                        message.entity_path = reader.string();
                        break;
                    }
                case 6: {
                        message.description = reader.string();
                        break;
                    }
                case 7: {
                        message.trashed = reader.bool();
                        break;
                    }
                case 8: {
                        message.entity_type_id = reader.string();
                        break;
                    }
                case 9: {
                        message.parent_id = reader.string();
                        break;
                    }
                case 10: {
                        message.preview_id = reader.string();
                        break;
                    }
                case 11: {
                        message.synced = reader.bool();
                        break;
                    }
                case 12: {
                        message.is_library = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an Entity message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Entity
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Entity} Entity
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Entity.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an Entity message.
         * @function verify
         * @memberof repository.Entity
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Entity.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.created_at != null && message.hasOwnProperty("created_at"))
                if (!$util.isString(message.created_at))
                    return "created_at: string expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.entity_path != null && message.hasOwnProperty("entity_path"))
                if (!$util.isString(message.entity_path))
                    return "entity_path: string expected";
            if (message.description != null && message.hasOwnProperty("description"))
                if (!$util.isString(message.description))
                    return "description: string expected";
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                if (typeof message.trashed !== "boolean")
                    return "trashed: boolean expected";
            if (message.entity_type_id != null && message.hasOwnProperty("entity_type_id"))
                if (!$util.isString(message.entity_type_id))
                    return "entity_type_id: string expected";
            if (message.parent_id != null && message.hasOwnProperty("parent_id"))
                if (!$util.isString(message.parent_id))
                    return "parent_id: string expected";
            if (message.preview_id != null && message.hasOwnProperty("preview_id"))
                if (!$util.isString(message.preview_id))
                    return "preview_id: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            if (message.is_library != null && message.hasOwnProperty("is_library"))
                if (typeof message.is_library !== "boolean")
                    return "is_library: boolean expected";
            return null;
        };

        /**
         * Creates an Entity message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Entity
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Entity} Entity
         */
        Entity.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Entity)
                return object;
            let message = new $root.repository.Entity();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.created_at != null)
                message.created_at = String(object.created_at);
            if (object.name != null)
                message.name = String(object.name);
            if (object.entity_path != null)
                message.entity_path = String(object.entity_path);
            if (object.description != null)
                message.description = String(object.description);
            if (object.trashed != null)
                message.trashed = Boolean(object.trashed);
            if (object.entity_type_id != null)
                message.entity_type_id = String(object.entity_type_id);
            if (object.parent_id != null)
                message.parent_id = String(object.parent_id);
            if (object.preview_id != null)
                message.preview_id = String(object.preview_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            if (object.is_library != null)
                message.is_library = Boolean(object.is_library);
            return message;
        };

        /**
         * Creates a plain object from an Entity message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Entity
         * @static
         * @param {repository.Entity} message Entity
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Entity.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.created_at = "";
                object.name = "";
                object.entity_path = "";
                object.description = "";
                object.trashed = false;
                object.entity_type_id = "";
                object.parent_id = "";
                object.preview_id = "";
                object.synced = false;
                object.is_library = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.created_at != null && message.hasOwnProperty("created_at"))
                object.created_at = message.created_at;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.entity_path != null && message.hasOwnProperty("entity_path"))
                object.entity_path = message.entity_path;
            if (message.description != null && message.hasOwnProperty("description"))
                object.description = message.description;
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                object.trashed = message.trashed;
            if (message.entity_type_id != null && message.hasOwnProperty("entity_type_id"))
                object.entity_type_id = message.entity_type_id;
            if (message.parent_id != null && message.hasOwnProperty("parent_id"))
                object.parent_id = message.parent_id;
            if (message.preview_id != null && message.hasOwnProperty("preview_id"))
                object.preview_id = message.preview_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            if (message.is_library != null && message.hasOwnProperty("is_library"))
                object.is_library = message.is_library;
            return object;
        };

        /**
         * Converts this Entity to JSON.
         * @function toJSON
         * @memberof repository.Entity
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Entity.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Entity
         * @function getTypeUrl
         * @memberof repository.Entity
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Entity.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Entity";
        };

        return Entity;
    })();

    repository.EntityAssignee = (function() {

        /**
         * Properties of an EntityAssignee.
         * @memberof repository
         * @interface IEntityAssignee
         * @property {string|null} [id] EntityAssignee id
         * @property {number|Long|null} [mtime] EntityAssignee mtime
         * @property {string|null} [entity_id] EntityAssignee entity_id
         * @property {string|null} [assignee_id] EntityAssignee assignee_id
         * @property {string|null} [assigner_id] EntityAssignee assigner_id
         * @property {boolean|null} [synced] EntityAssignee synced
         */

        /**
         * Constructs a new EntityAssignee.
         * @memberof repository
         * @classdesc Represents an EntityAssignee.
         * @implements IEntityAssignee
         * @constructor
         * @param {repository.IEntityAssignee=} [properties] Properties to set
         */
        function EntityAssignee(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * EntityAssignee id.
         * @member {string} id
         * @memberof repository.EntityAssignee
         * @instance
         */
        EntityAssignee.prototype.id = "";

        /**
         * EntityAssignee mtime.
         * @member {number|Long} mtime
         * @memberof repository.EntityAssignee
         * @instance
         */
        EntityAssignee.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * EntityAssignee entity_id.
         * @member {string} entity_id
         * @memberof repository.EntityAssignee
         * @instance
         */
        EntityAssignee.prototype.entity_id = "";

        /**
         * EntityAssignee assignee_id.
         * @member {string} assignee_id
         * @memberof repository.EntityAssignee
         * @instance
         */
        EntityAssignee.prototype.assignee_id = "";

        /**
         * EntityAssignee assigner_id.
         * @member {string} assigner_id
         * @memberof repository.EntityAssignee
         * @instance
         */
        EntityAssignee.prototype.assigner_id = "";

        /**
         * EntityAssignee synced.
         * @member {boolean} synced
         * @memberof repository.EntityAssignee
         * @instance
         */
        EntityAssignee.prototype.synced = false;

        /**
         * Creates a new EntityAssignee instance using the specified properties.
         * @function create
         * @memberof repository.EntityAssignee
         * @static
         * @param {repository.IEntityAssignee=} [properties] Properties to set
         * @returns {repository.EntityAssignee} EntityAssignee instance
         */
        EntityAssignee.create = function create(properties) {
            return new EntityAssignee(properties);
        };

        /**
         * Encodes the specified EntityAssignee message. Does not implicitly {@link repository.EntityAssignee.verify|verify} messages.
         * @function encode
         * @memberof repository.EntityAssignee
         * @static
         * @param {repository.IEntityAssignee} message EntityAssignee message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EntityAssignee.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.entity_id != null && Object.hasOwnProperty.call(message, "entity_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.entity_id);
            if (message.assignee_id != null && Object.hasOwnProperty.call(message, "assignee_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.assignee_id);
            if (message.assigner_id != null && Object.hasOwnProperty.call(message, "assigner_id"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.assigner_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified EntityAssignee message, length delimited. Does not implicitly {@link repository.EntityAssignee.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.EntityAssignee
         * @static
         * @param {repository.IEntityAssignee} message EntityAssignee message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EntityAssignee.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an EntityAssignee message from the specified reader or buffer.
         * @function decode
         * @memberof repository.EntityAssignee
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.EntityAssignee} EntityAssignee
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EntityAssignee.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.EntityAssignee();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.entity_id = reader.string();
                        break;
                    }
                case 4: {
                        message.assignee_id = reader.string();
                        break;
                    }
                case 5: {
                        message.assigner_id = reader.string();
                        break;
                    }
                case 6: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an EntityAssignee message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.EntityAssignee
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.EntityAssignee} EntityAssignee
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EntityAssignee.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an EntityAssignee message.
         * @function verify
         * @memberof repository.EntityAssignee
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        EntityAssignee.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.entity_id != null && message.hasOwnProperty("entity_id"))
                if (!$util.isString(message.entity_id))
                    return "entity_id: string expected";
            if (message.assignee_id != null && message.hasOwnProperty("assignee_id"))
                if (!$util.isString(message.assignee_id))
                    return "assignee_id: string expected";
            if (message.assigner_id != null && message.hasOwnProperty("assigner_id"))
                if (!$util.isString(message.assigner_id))
                    return "assigner_id: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates an EntityAssignee message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.EntityAssignee
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.EntityAssignee} EntityAssignee
         */
        EntityAssignee.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.EntityAssignee)
                return object;
            let message = new $root.repository.EntityAssignee();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.entity_id != null)
                message.entity_id = String(object.entity_id);
            if (object.assignee_id != null)
                message.assignee_id = String(object.assignee_id);
            if (object.assigner_id != null)
                message.assigner_id = String(object.assigner_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from an EntityAssignee message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.EntityAssignee
         * @static
         * @param {repository.EntityAssignee} message EntityAssignee
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        EntityAssignee.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.entity_id = "";
                object.assignee_id = "";
                object.assigner_id = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.entity_id != null && message.hasOwnProperty("entity_id"))
                object.entity_id = message.entity_id;
            if (message.assignee_id != null && message.hasOwnProperty("assignee_id"))
                object.assignee_id = message.assignee_id;
            if (message.assigner_id != null && message.hasOwnProperty("assigner_id"))
                object.assigner_id = message.assigner_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this EntityAssignee to JSON.
         * @function toJSON
         * @memberof repository.EntityAssignee
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        EntityAssignee.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for EntityAssignee
         * @function getTypeUrl
         * @memberof repository.EntityAssignee
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        EntityAssignee.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.EntityAssignee";
        };

        return EntityAssignee;
    })();

    repository.TaskDependency = (function() {

        /**
         * Properties of a TaskDependency.
         * @memberof repository
         * @interface ITaskDependency
         * @property {string|null} [id] TaskDependency id
         * @property {number|Long|null} [mtime] TaskDependency mtime
         * @property {string|null} [task_id] TaskDependency task_id
         * @property {string|null} [dependency_id] TaskDependency dependency_id
         * @property {string|null} [dependency_type_id] TaskDependency dependency_type_id
         * @property {boolean|null} [synced] TaskDependency synced
         */

        /**
         * Constructs a new TaskDependency.
         * @memberof repository
         * @classdesc Represents a TaskDependency.
         * @implements ITaskDependency
         * @constructor
         * @param {repository.ITaskDependency=} [properties] Properties to set
         */
        function TaskDependency(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * TaskDependency id.
         * @member {string} id
         * @memberof repository.TaskDependency
         * @instance
         */
        TaskDependency.prototype.id = "";

        /**
         * TaskDependency mtime.
         * @member {number|Long} mtime
         * @memberof repository.TaskDependency
         * @instance
         */
        TaskDependency.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TaskDependency task_id.
         * @member {string} task_id
         * @memberof repository.TaskDependency
         * @instance
         */
        TaskDependency.prototype.task_id = "";

        /**
         * TaskDependency dependency_id.
         * @member {string} dependency_id
         * @memberof repository.TaskDependency
         * @instance
         */
        TaskDependency.prototype.dependency_id = "";

        /**
         * TaskDependency dependency_type_id.
         * @member {string} dependency_type_id
         * @memberof repository.TaskDependency
         * @instance
         */
        TaskDependency.prototype.dependency_type_id = "";

        /**
         * TaskDependency synced.
         * @member {boolean} synced
         * @memberof repository.TaskDependency
         * @instance
         */
        TaskDependency.prototype.synced = false;

        /**
         * Creates a new TaskDependency instance using the specified properties.
         * @function create
         * @memberof repository.TaskDependency
         * @static
         * @param {repository.ITaskDependency=} [properties] Properties to set
         * @returns {repository.TaskDependency} TaskDependency instance
         */
        TaskDependency.create = function create(properties) {
            return new TaskDependency(properties);
        };

        /**
         * Encodes the specified TaskDependency message. Does not implicitly {@link repository.TaskDependency.verify|verify} messages.
         * @function encode
         * @memberof repository.TaskDependency
         * @static
         * @param {repository.ITaskDependency} message TaskDependency message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TaskDependency.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.task_id != null && Object.hasOwnProperty.call(message, "task_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.task_id);
            if (message.dependency_id != null && Object.hasOwnProperty.call(message, "dependency_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.dependency_id);
            if (message.dependency_type_id != null && Object.hasOwnProperty.call(message, "dependency_type_id"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.dependency_type_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified TaskDependency message, length delimited. Does not implicitly {@link repository.TaskDependency.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.TaskDependency
         * @static
         * @param {repository.ITaskDependency} message TaskDependency message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TaskDependency.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a TaskDependency message from the specified reader or buffer.
         * @function decode
         * @memberof repository.TaskDependency
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.TaskDependency} TaskDependency
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TaskDependency.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.TaskDependency();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.task_id = reader.string();
                        break;
                    }
                case 4: {
                        message.dependency_id = reader.string();
                        break;
                    }
                case 5: {
                        message.dependency_type_id = reader.string();
                        break;
                    }
                case 6: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a TaskDependency message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.TaskDependency
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.TaskDependency} TaskDependency
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TaskDependency.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a TaskDependency message.
         * @function verify
         * @memberof repository.TaskDependency
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        TaskDependency.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.task_id != null && message.hasOwnProperty("task_id"))
                if (!$util.isString(message.task_id))
                    return "task_id: string expected";
            if (message.dependency_id != null && message.hasOwnProperty("dependency_id"))
                if (!$util.isString(message.dependency_id))
                    return "dependency_id: string expected";
            if (message.dependency_type_id != null && message.hasOwnProperty("dependency_type_id"))
                if (!$util.isString(message.dependency_type_id))
                    return "dependency_type_id: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a TaskDependency message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.TaskDependency
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.TaskDependency} TaskDependency
         */
        TaskDependency.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.TaskDependency)
                return object;
            let message = new $root.repository.TaskDependency();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.task_id != null)
                message.task_id = String(object.task_id);
            if (object.dependency_id != null)
                message.dependency_id = String(object.dependency_id);
            if (object.dependency_type_id != null)
                message.dependency_type_id = String(object.dependency_type_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a TaskDependency message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.TaskDependency
         * @static
         * @param {repository.TaskDependency} message TaskDependency
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        TaskDependency.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.task_id = "";
                object.dependency_id = "";
                object.dependency_type_id = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.task_id != null && message.hasOwnProperty("task_id"))
                object.task_id = message.task_id;
            if (message.dependency_id != null && message.hasOwnProperty("dependency_id"))
                object.dependency_id = message.dependency_id;
            if (message.dependency_type_id != null && message.hasOwnProperty("dependency_type_id"))
                object.dependency_type_id = message.dependency_type_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this TaskDependency to JSON.
         * @function toJSON
         * @memberof repository.TaskDependency
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        TaskDependency.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for TaskDependency
         * @function getTypeUrl
         * @memberof repository.TaskDependency
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        TaskDependency.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.TaskDependency";
        };

        return TaskDependency;
    })();

    repository.EntityDependency = (function() {

        /**
         * Properties of an EntityDependency.
         * @memberof repository
         * @interface IEntityDependency
         * @property {string|null} [id] EntityDependency id
         * @property {number|Long|null} [mtime] EntityDependency mtime
         * @property {string|null} [task_id] EntityDependency task_id
         * @property {string|null} [dependency_id] EntityDependency dependency_id
         * @property {string|null} [dependency_type_id] EntityDependency dependency_type_id
         * @property {boolean|null} [synced] EntityDependency synced
         */

        /**
         * Constructs a new EntityDependency.
         * @memberof repository
         * @classdesc Represents an EntityDependency.
         * @implements IEntityDependency
         * @constructor
         * @param {repository.IEntityDependency=} [properties] Properties to set
         */
        function EntityDependency(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * EntityDependency id.
         * @member {string} id
         * @memberof repository.EntityDependency
         * @instance
         */
        EntityDependency.prototype.id = "";

        /**
         * EntityDependency mtime.
         * @member {number|Long} mtime
         * @memberof repository.EntityDependency
         * @instance
         */
        EntityDependency.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * EntityDependency task_id.
         * @member {string} task_id
         * @memberof repository.EntityDependency
         * @instance
         */
        EntityDependency.prototype.task_id = "";

        /**
         * EntityDependency dependency_id.
         * @member {string} dependency_id
         * @memberof repository.EntityDependency
         * @instance
         */
        EntityDependency.prototype.dependency_id = "";

        /**
         * EntityDependency dependency_type_id.
         * @member {string} dependency_type_id
         * @memberof repository.EntityDependency
         * @instance
         */
        EntityDependency.prototype.dependency_type_id = "";

        /**
         * EntityDependency synced.
         * @member {boolean} synced
         * @memberof repository.EntityDependency
         * @instance
         */
        EntityDependency.prototype.synced = false;

        /**
         * Creates a new EntityDependency instance using the specified properties.
         * @function create
         * @memberof repository.EntityDependency
         * @static
         * @param {repository.IEntityDependency=} [properties] Properties to set
         * @returns {repository.EntityDependency} EntityDependency instance
         */
        EntityDependency.create = function create(properties) {
            return new EntityDependency(properties);
        };

        /**
         * Encodes the specified EntityDependency message. Does not implicitly {@link repository.EntityDependency.verify|verify} messages.
         * @function encode
         * @memberof repository.EntityDependency
         * @static
         * @param {repository.IEntityDependency} message EntityDependency message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EntityDependency.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.task_id != null && Object.hasOwnProperty.call(message, "task_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.task_id);
            if (message.dependency_id != null && Object.hasOwnProperty.call(message, "dependency_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.dependency_id);
            if (message.dependency_type_id != null && Object.hasOwnProperty.call(message, "dependency_type_id"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.dependency_type_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified EntityDependency message, length delimited. Does not implicitly {@link repository.EntityDependency.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.EntityDependency
         * @static
         * @param {repository.IEntityDependency} message EntityDependency message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        EntityDependency.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an EntityDependency message from the specified reader or buffer.
         * @function decode
         * @memberof repository.EntityDependency
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.EntityDependency} EntityDependency
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EntityDependency.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.EntityDependency();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.task_id = reader.string();
                        break;
                    }
                case 4: {
                        message.dependency_id = reader.string();
                        break;
                    }
                case 5: {
                        message.dependency_type_id = reader.string();
                        break;
                    }
                case 6: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an EntityDependency message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.EntityDependency
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.EntityDependency} EntityDependency
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        EntityDependency.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an EntityDependency message.
         * @function verify
         * @memberof repository.EntityDependency
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        EntityDependency.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.task_id != null && message.hasOwnProperty("task_id"))
                if (!$util.isString(message.task_id))
                    return "task_id: string expected";
            if (message.dependency_id != null && message.hasOwnProperty("dependency_id"))
                if (!$util.isString(message.dependency_id))
                    return "dependency_id: string expected";
            if (message.dependency_type_id != null && message.hasOwnProperty("dependency_type_id"))
                if (!$util.isString(message.dependency_type_id))
                    return "dependency_type_id: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates an EntityDependency message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.EntityDependency
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.EntityDependency} EntityDependency
         */
        EntityDependency.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.EntityDependency)
                return object;
            let message = new $root.repository.EntityDependency();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.task_id != null)
                message.task_id = String(object.task_id);
            if (object.dependency_id != null)
                message.dependency_id = String(object.dependency_id);
            if (object.dependency_type_id != null)
                message.dependency_type_id = String(object.dependency_type_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from an EntityDependency message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.EntityDependency
         * @static
         * @param {repository.EntityDependency} message EntityDependency
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        EntityDependency.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.task_id = "";
                object.dependency_id = "";
                object.dependency_type_id = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.task_id != null && message.hasOwnProperty("task_id"))
                object.task_id = message.task_id;
            if (message.dependency_id != null && message.hasOwnProperty("dependency_id"))
                object.dependency_id = message.dependency_id;
            if (message.dependency_type_id != null && message.hasOwnProperty("dependency_type_id"))
                object.dependency_type_id = message.dependency_type_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this EntityDependency to JSON.
         * @function toJSON
         * @memberof repository.EntityDependency
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        EntityDependency.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for EntityDependency
         * @function getTypeUrl
         * @memberof repository.EntityDependency
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        EntityDependency.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.EntityDependency";
        };

        return EntityDependency;
    })();

    repository.Workflow = (function() {

        /**
         * Properties of a Workflow.
         * @memberof repository
         * @interface IWorkflow
         * @property {string|null} [id] Workflow id
         * @property {number|Long|null} [mtime] Workflow mtime
         * @property {string|null} [name] Workflow name
         * @property {boolean|null} [synced] Workflow synced
         */

        /**
         * Constructs a new Workflow.
         * @memberof repository
         * @classdesc Represents a Workflow.
         * @implements IWorkflow
         * @constructor
         * @param {repository.IWorkflow=} [properties] Properties to set
         */
        function Workflow(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Workflow id.
         * @member {string} id
         * @memberof repository.Workflow
         * @instance
         */
        Workflow.prototype.id = "";

        /**
         * Workflow mtime.
         * @member {number|Long} mtime
         * @memberof repository.Workflow
         * @instance
         */
        Workflow.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Workflow name.
         * @member {string} name
         * @memberof repository.Workflow
         * @instance
         */
        Workflow.prototype.name = "";

        /**
         * Workflow synced.
         * @member {boolean} synced
         * @memberof repository.Workflow
         * @instance
         */
        Workflow.prototype.synced = false;

        /**
         * Creates a new Workflow instance using the specified properties.
         * @function create
         * @memberof repository.Workflow
         * @static
         * @param {repository.IWorkflow=} [properties] Properties to set
         * @returns {repository.Workflow} Workflow instance
         */
        Workflow.create = function create(properties) {
            return new Workflow(properties);
        };

        /**
         * Encodes the specified Workflow message. Does not implicitly {@link repository.Workflow.verify|verify} messages.
         * @function encode
         * @memberof repository.Workflow
         * @static
         * @param {repository.IWorkflow} message Workflow message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Workflow.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 4, wireType 0 =*/32).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified Workflow message, length delimited. Does not implicitly {@link repository.Workflow.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Workflow
         * @static
         * @param {repository.IWorkflow} message Workflow message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Workflow.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Workflow message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Workflow
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Workflow} Workflow
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Workflow.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Workflow();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Workflow message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Workflow
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Workflow} Workflow
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Workflow.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Workflow message.
         * @function verify
         * @memberof repository.Workflow
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Workflow.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a Workflow message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Workflow
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Workflow} Workflow
         */
        Workflow.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Workflow)
                return object;
            let message = new $root.repository.Workflow();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a Workflow message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Workflow
         * @static
         * @param {repository.Workflow} message Workflow
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Workflow.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this Workflow to JSON.
         * @function toJSON
         * @memberof repository.Workflow
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Workflow.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Workflow
         * @function getTypeUrl
         * @memberof repository.Workflow
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Workflow.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Workflow";
        };

        return Workflow;
    })();

    repository.WorkflowTask = (function() {

        /**
         * Properties of a WorkflowTask.
         * @memberof repository
         * @interface IWorkflowTask
         * @property {string|null} [id] WorkflowTask id
         * @property {number|Long|null} [mtime] WorkflowTask mtime
         * @property {string|null} [name] WorkflowTask name
         * @property {string|null} [template_id] WorkflowTask template_id
         * @property {boolean|null} [is_resource] WorkflowTask is_resource
         * @property {string|null} [workflow_id] WorkflowTask workflow_id
         * @property {string|null} [task_type_id] WorkflowTask task_type_id
         * @property {string|null} [workflow_entity_id] WorkflowTask workflow_entity_id
         * @property {boolean|null} [is_link] WorkflowTask is_link
         * @property {string|null} [pointer] WorkflowTask pointer
         * @property {boolean|null} [synced] WorkflowTask synced
         */

        /**
         * Constructs a new WorkflowTask.
         * @memberof repository
         * @classdesc Represents a WorkflowTask.
         * @implements IWorkflowTask
         * @constructor
         * @param {repository.IWorkflowTask=} [properties] Properties to set
         */
        function WorkflowTask(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WorkflowTask id.
         * @member {string} id
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.id = "";

        /**
         * WorkflowTask mtime.
         * @member {number|Long} mtime
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * WorkflowTask name.
         * @member {string} name
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.name = "";

        /**
         * WorkflowTask template_id.
         * @member {string} template_id
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.template_id = "";

        /**
         * WorkflowTask is_resource.
         * @member {boolean} is_resource
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.is_resource = false;

        /**
         * WorkflowTask workflow_id.
         * @member {string} workflow_id
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.workflow_id = "";

        /**
         * WorkflowTask task_type_id.
         * @member {string} task_type_id
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.task_type_id = "";

        /**
         * WorkflowTask workflow_entity_id.
         * @member {string} workflow_entity_id
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.workflow_entity_id = "";

        /**
         * WorkflowTask is_link.
         * @member {boolean} is_link
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.is_link = false;

        /**
         * WorkflowTask pointer.
         * @member {string} pointer
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.pointer = "";

        /**
         * WorkflowTask synced.
         * @member {boolean} synced
         * @memberof repository.WorkflowTask
         * @instance
         */
        WorkflowTask.prototype.synced = false;

        /**
         * Creates a new WorkflowTask instance using the specified properties.
         * @function create
         * @memberof repository.WorkflowTask
         * @static
         * @param {repository.IWorkflowTask=} [properties] Properties to set
         * @returns {repository.WorkflowTask} WorkflowTask instance
         */
        WorkflowTask.create = function create(properties) {
            return new WorkflowTask(properties);
        };

        /**
         * Encodes the specified WorkflowTask message. Does not implicitly {@link repository.WorkflowTask.verify|verify} messages.
         * @function encode
         * @memberof repository.WorkflowTask
         * @static
         * @param {repository.IWorkflowTask} message WorkflowTask message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WorkflowTask.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.template_id != null && Object.hasOwnProperty.call(message, "template_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.template_id);
            if (message.is_resource != null && Object.hasOwnProperty.call(message, "is_resource"))
                writer.uint32(/* id 5, wireType 0 =*/40).bool(message.is_resource);
            if (message.workflow_id != null && Object.hasOwnProperty.call(message, "workflow_id"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.workflow_id);
            if (message.task_type_id != null && Object.hasOwnProperty.call(message, "task_type_id"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.task_type_id);
            if (message.workflow_entity_id != null && Object.hasOwnProperty.call(message, "workflow_entity_id"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.workflow_entity_id);
            if (message.is_link != null && Object.hasOwnProperty.call(message, "is_link"))
                writer.uint32(/* id 9, wireType 0 =*/72).bool(message.is_link);
            if (message.pointer != null && Object.hasOwnProperty.call(message, "pointer"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.pointer);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 11, wireType 0 =*/88).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified WorkflowTask message, length delimited. Does not implicitly {@link repository.WorkflowTask.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.WorkflowTask
         * @static
         * @param {repository.IWorkflowTask} message WorkflowTask message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WorkflowTask.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a WorkflowTask message from the specified reader or buffer.
         * @function decode
         * @memberof repository.WorkflowTask
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.WorkflowTask} WorkflowTask
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WorkflowTask.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.WorkflowTask();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.template_id = reader.string();
                        break;
                    }
                case 5: {
                        message.is_resource = reader.bool();
                        break;
                    }
                case 6: {
                        message.workflow_id = reader.string();
                        break;
                    }
                case 7: {
                        message.task_type_id = reader.string();
                        break;
                    }
                case 8: {
                        message.workflow_entity_id = reader.string();
                        break;
                    }
                case 9: {
                        message.is_link = reader.bool();
                        break;
                    }
                case 10: {
                        message.pointer = reader.string();
                        break;
                    }
                case 11: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a WorkflowTask message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.WorkflowTask
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.WorkflowTask} WorkflowTask
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WorkflowTask.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a WorkflowTask message.
         * @function verify
         * @memberof repository.WorkflowTask
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        WorkflowTask.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.template_id != null && message.hasOwnProperty("template_id"))
                if (!$util.isString(message.template_id))
                    return "template_id: string expected";
            if (message.is_resource != null && message.hasOwnProperty("is_resource"))
                if (typeof message.is_resource !== "boolean")
                    return "is_resource: boolean expected";
            if (message.workflow_id != null && message.hasOwnProperty("workflow_id"))
                if (!$util.isString(message.workflow_id))
                    return "workflow_id: string expected";
            if (message.task_type_id != null && message.hasOwnProperty("task_type_id"))
                if (!$util.isString(message.task_type_id))
                    return "task_type_id: string expected";
            if (message.workflow_entity_id != null && message.hasOwnProperty("workflow_entity_id"))
                if (!$util.isString(message.workflow_entity_id))
                    return "workflow_entity_id: string expected";
            if (message.is_link != null && message.hasOwnProperty("is_link"))
                if (typeof message.is_link !== "boolean")
                    return "is_link: boolean expected";
            if (message.pointer != null && message.hasOwnProperty("pointer"))
                if (!$util.isString(message.pointer))
                    return "pointer: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a WorkflowTask message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.WorkflowTask
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.WorkflowTask} WorkflowTask
         */
        WorkflowTask.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.WorkflowTask)
                return object;
            let message = new $root.repository.WorkflowTask();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.template_id != null)
                message.template_id = String(object.template_id);
            if (object.is_resource != null)
                message.is_resource = Boolean(object.is_resource);
            if (object.workflow_id != null)
                message.workflow_id = String(object.workflow_id);
            if (object.task_type_id != null)
                message.task_type_id = String(object.task_type_id);
            if (object.workflow_entity_id != null)
                message.workflow_entity_id = String(object.workflow_entity_id);
            if (object.is_link != null)
                message.is_link = Boolean(object.is_link);
            if (object.pointer != null)
                message.pointer = String(object.pointer);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a WorkflowTask message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.WorkflowTask
         * @static
         * @param {repository.WorkflowTask} message WorkflowTask
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        WorkflowTask.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.template_id = "";
                object.is_resource = false;
                object.workflow_id = "";
                object.task_type_id = "";
                object.workflow_entity_id = "";
                object.is_link = false;
                object.pointer = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.template_id != null && message.hasOwnProperty("template_id"))
                object.template_id = message.template_id;
            if (message.is_resource != null && message.hasOwnProperty("is_resource"))
                object.is_resource = message.is_resource;
            if (message.workflow_id != null && message.hasOwnProperty("workflow_id"))
                object.workflow_id = message.workflow_id;
            if (message.task_type_id != null && message.hasOwnProperty("task_type_id"))
                object.task_type_id = message.task_type_id;
            if (message.workflow_entity_id != null && message.hasOwnProperty("workflow_entity_id"))
                object.workflow_entity_id = message.workflow_entity_id;
            if (message.is_link != null && message.hasOwnProperty("is_link"))
                object.is_link = message.is_link;
            if (message.pointer != null && message.hasOwnProperty("pointer"))
                object.pointer = message.pointer;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this WorkflowTask to JSON.
         * @function toJSON
         * @memberof repository.WorkflowTask
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        WorkflowTask.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for WorkflowTask
         * @function getTypeUrl
         * @memberof repository.WorkflowTask
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WorkflowTask.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.WorkflowTask";
        };

        return WorkflowTask;
    })();

    repository.WorkflowEntity = (function() {

        /**
         * Properties of a WorkflowEntity.
         * @memberof repository
         * @interface IWorkflowEntity
         * @property {string|null} [id] WorkflowEntity id
         * @property {number|Long|null} [mtime] WorkflowEntity mtime
         * @property {string|null} [name] WorkflowEntity name
         * @property {string|null} [workflow_id] WorkflowEntity workflow_id
         * @property {string|null} [entity_type_id] WorkflowEntity entity_type_id
         * @property {string|null} [parent_id] WorkflowEntity parent_id
         * @property {boolean|null} [synced] WorkflowEntity synced
         */

        /**
         * Constructs a new WorkflowEntity.
         * @memberof repository
         * @classdesc Represents a WorkflowEntity.
         * @implements IWorkflowEntity
         * @constructor
         * @param {repository.IWorkflowEntity=} [properties] Properties to set
         */
        function WorkflowEntity(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WorkflowEntity id.
         * @member {string} id
         * @memberof repository.WorkflowEntity
         * @instance
         */
        WorkflowEntity.prototype.id = "";

        /**
         * WorkflowEntity mtime.
         * @member {number|Long} mtime
         * @memberof repository.WorkflowEntity
         * @instance
         */
        WorkflowEntity.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * WorkflowEntity name.
         * @member {string} name
         * @memberof repository.WorkflowEntity
         * @instance
         */
        WorkflowEntity.prototype.name = "";

        /**
         * WorkflowEntity workflow_id.
         * @member {string} workflow_id
         * @memberof repository.WorkflowEntity
         * @instance
         */
        WorkflowEntity.prototype.workflow_id = "";

        /**
         * WorkflowEntity entity_type_id.
         * @member {string} entity_type_id
         * @memberof repository.WorkflowEntity
         * @instance
         */
        WorkflowEntity.prototype.entity_type_id = "";

        /**
         * WorkflowEntity parent_id.
         * @member {string} parent_id
         * @memberof repository.WorkflowEntity
         * @instance
         */
        WorkflowEntity.prototype.parent_id = "";

        /**
         * WorkflowEntity synced.
         * @member {boolean} synced
         * @memberof repository.WorkflowEntity
         * @instance
         */
        WorkflowEntity.prototype.synced = false;

        /**
         * Creates a new WorkflowEntity instance using the specified properties.
         * @function create
         * @memberof repository.WorkflowEntity
         * @static
         * @param {repository.IWorkflowEntity=} [properties] Properties to set
         * @returns {repository.WorkflowEntity} WorkflowEntity instance
         */
        WorkflowEntity.create = function create(properties) {
            return new WorkflowEntity(properties);
        };

        /**
         * Encodes the specified WorkflowEntity message. Does not implicitly {@link repository.WorkflowEntity.verify|verify} messages.
         * @function encode
         * @memberof repository.WorkflowEntity
         * @static
         * @param {repository.IWorkflowEntity} message WorkflowEntity message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WorkflowEntity.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.workflow_id != null && Object.hasOwnProperty.call(message, "workflow_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.workflow_id);
            if (message.entity_type_id != null && Object.hasOwnProperty.call(message, "entity_type_id"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.entity_type_id);
            if (message.parent_id != null && Object.hasOwnProperty.call(message, "parent_id"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.parent_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 7, wireType 0 =*/56).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified WorkflowEntity message, length delimited. Does not implicitly {@link repository.WorkflowEntity.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.WorkflowEntity
         * @static
         * @param {repository.IWorkflowEntity} message WorkflowEntity message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WorkflowEntity.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a WorkflowEntity message from the specified reader or buffer.
         * @function decode
         * @memberof repository.WorkflowEntity
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.WorkflowEntity} WorkflowEntity
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WorkflowEntity.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.WorkflowEntity();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.workflow_id = reader.string();
                        break;
                    }
                case 5: {
                        message.entity_type_id = reader.string();
                        break;
                    }
                case 6: {
                        message.parent_id = reader.string();
                        break;
                    }
                case 7: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a WorkflowEntity message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.WorkflowEntity
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.WorkflowEntity} WorkflowEntity
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WorkflowEntity.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a WorkflowEntity message.
         * @function verify
         * @memberof repository.WorkflowEntity
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        WorkflowEntity.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.workflow_id != null && message.hasOwnProperty("workflow_id"))
                if (!$util.isString(message.workflow_id))
                    return "workflow_id: string expected";
            if (message.entity_type_id != null && message.hasOwnProperty("entity_type_id"))
                if (!$util.isString(message.entity_type_id))
                    return "entity_type_id: string expected";
            if (message.parent_id != null && message.hasOwnProperty("parent_id"))
                if (!$util.isString(message.parent_id))
                    return "parent_id: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a WorkflowEntity message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.WorkflowEntity
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.WorkflowEntity} WorkflowEntity
         */
        WorkflowEntity.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.WorkflowEntity)
                return object;
            let message = new $root.repository.WorkflowEntity();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.workflow_id != null)
                message.workflow_id = String(object.workflow_id);
            if (object.entity_type_id != null)
                message.entity_type_id = String(object.entity_type_id);
            if (object.parent_id != null)
                message.parent_id = String(object.parent_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a WorkflowEntity message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.WorkflowEntity
         * @static
         * @param {repository.WorkflowEntity} message WorkflowEntity
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        WorkflowEntity.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.workflow_id = "";
                object.entity_type_id = "";
                object.parent_id = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.workflow_id != null && message.hasOwnProperty("workflow_id"))
                object.workflow_id = message.workflow_id;
            if (message.entity_type_id != null && message.hasOwnProperty("entity_type_id"))
                object.entity_type_id = message.entity_type_id;
            if (message.parent_id != null && message.hasOwnProperty("parent_id"))
                object.parent_id = message.parent_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this WorkflowEntity to JSON.
         * @function toJSON
         * @memberof repository.WorkflowEntity
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        WorkflowEntity.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for WorkflowEntity
         * @function getTypeUrl
         * @memberof repository.WorkflowEntity
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WorkflowEntity.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.WorkflowEntity";
        };

        return WorkflowEntity;
    })();

    repository.WorkflowLink = (function() {

        /**
         * Properties of a WorkflowLink.
         * @memberof repository
         * @interface IWorkflowLink
         * @property {string|null} [id] WorkflowLink id
         * @property {number|Long|null} [mtime] WorkflowLink mtime
         * @property {string|null} [name] WorkflowLink name
         * @property {string|null} [entity_type_id] WorkflowLink entity_type_id
         * @property {string|null} [workflow_id] WorkflowLink workflow_id
         * @property {string|null} [linked_workflow_id] WorkflowLink linked_workflow_id
         * @property {string|null} [linked_workflow_name] WorkflowLink linked_workflow_name
         * @property {boolean|null} [synced] WorkflowLink synced
         */

        /**
         * Constructs a new WorkflowLink.
         * @memberof repository
         * @classdesc Represents a WorkflowLink.
         * @implements IWorkflowLink
         * @constructor
         * @param {repository.IWorkflowLink=} [properties] Properties to set
         */
        function WorkflowLink(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WorkflowLink id.
         * @member {string} id
         * @memberof repository.WorkflowLink
         * @instance
         */
        WorkflowLink.prototype.id = "";

        /**
         * WorkflowLink mtime.
         * @member {number|Long} mtime
         * @memberof repository.WorkflowLink
         * @instance
         */
        WorkflowLink.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * WorkflowLink name.
         * @member {string} name
         * @memberof repository.WorkflowLink
         * @instance
         */
        WorkflowLink.prototype.name = "";

        /**
         * WorkflowLink entity_type_id.
         * @member {string} entity_type_id
         * @memberof repository.WorkflowLink
         * @instance
         */
        WorkflowLink.prototype.entity_type_id = "";

        /**
         * WorkflowLink workflow_id.
         * @member {string} workflow_id
         * @memberof repository.WorkflowLink
         * @instance
         */
        WorkflowLink.prototype.workflow_id = "";

        /**
         * WorkflowLink linked_workflow_id.
         * @member {string} linked_workflow_id
         * @memberof repository.WorkflowLink
         * @instance
         */
        WorkflowLink.prototype.linked_workflow_id = "";

        /**
         * WorkflowLink linked_workflow_name.
         * @member {string} linked_workflow_name
         * @memberof repository.WorkflowLink
         * @instance
         */
        WorkflowLink.prototype.linked_workflow_name = "";

        /**
         * WorkflowLink synced.
         * @member {boolean} synced
         * @memberof repository.WorkflowLink
         * @instance
         */
        WorkflowLink.prototype.synced = false;

        /**
         * Creates a new WorkflowLink instance using the specified properties.
         * @function create
         * @memberof repository.WorkflowLink
         * @static
         * @param {repository.IWorkflowLink=} [properties] Properties to set
         * @returns {repository.WorkflowLink} WorkflowLink instance
         */
        WorkflowLink.create = function create(properties) {
            return new WorkflowLink(properties);
        };

        /**
         * Encodes the specified WorkflowLink message. Does not implicitly {@link repository.WorkflowLink.verify|verify} messages.
         * @function encode
         * @memberof repository.WorkflowLink
         * @static
         * @param {repository.IWorkflowLink} message WorkflowLink message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WorkflowLink.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.entity_type_id != null && Object.hasOwnProperty.call(message, "entity_type_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.entity_type_id);
            if (message.workflow_id != null && Object.hasOwnProperty.call(message, "workflow_id"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.workflow_id);
            if (message.linked_workflow_id != null && Object.hasOwnProperty.call(message, "linked_workflow_id"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.linked_workflow_id);
            if (message.linked_workflow_name != null && Object.hasOwnProperty.call(message, "linked_workflow_name"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.linked_workflow_name);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 8, wireType 0 =*/64).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified WorkflowLink message, length delimited. Does not implicitly {@link repository.WorkflowLink.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.WorkflowLink
         * @static
         * @param {repository.IWorkflowLink} message WorkflowLink message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WorkflowLink.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a WorkflowLink message from the specified reader or buffer.
         * @function decode
         * @memberof repository.WorkflowLink
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.WorkflowLink} WorkflowLink
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WorkflowLink.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.WorkflowLink();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.entity_type_id = reader.string();
                        break;
                    }
                case 5: {
                        message.workflow_id = reader.string();
                        break;
                    }
                case 6: {
                        message.linked_workflow_id = reader.string();
                        break;
                    }
                case 7: {
                        message.linked_workflow_name = reader.string();
                        break;
                    }
                case 8: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a WorkflowLink message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.WorkflowLink
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.WorkflowLink} WorkflowLink
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WorkflowLink.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a WorkflowLink message.
         * @function verify
         * @memberof repository.WorkflowLink
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        WorkflowLink.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.entity_type_id != null && message.hasOwnProperty("entity_type_id"))
                if (!$util.isString(message.entity_type_id))
                    return "entity_type_id: string expected";
            if (message.workflow_id != null && message.hasOwnProperty("workflow_id"))
                if (!$util.isString(message.workflow_id))
                    return "workflow_id: string expected";
            if (message.linked_workflow_id != null && message.hasOwnProperty("linked_workflow_id"))
                if (!$util.isString(message.linked_workflow_id))
                    return "linked_workflow_id: string expected";
            if (message.linked_workflow_name != null && message.hasOwnProperty("linked_workflow_name"))
                if (!$util.isString(message.linked_workflow_name))
                    return "linked_workflow_name: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a WorkflowLink message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.WorkflowLink
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.WorkflowLink} WorkflowLink
         */
        WorkflowLink.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.WorkflowLink)
                return object;
            let message = new $root.repository.WorkflowLink();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.entity_type_id != null)
                message.entity_type_id = String(object.entity_type_id);
            if (object.workflow_id != null)
                message.workflow_id = String(object.workflow_id);
            if (object.linked_workflow_id != null)
                message.linked_workflow_id = String(object.linked_workflow_id);
            if (object.linked_workflow_name != null)
                message.linked_workflow_name = String(object.linked_workflow_name);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a WorkflowLink message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.WorkflowLink
         * @static
         * @param {repository.WorkflowLink} message WorkflowLink
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        WorkflowLink.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.entity_type_id = "";
                object.workflow_id = "";
                object.linked_workflow_id = "";
                object.linked_workflow_name = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.entity_type_id != null && message.hasOwnProperty("entity_type_id"))
                object.entity_type_id = message.entity_type_id;
            if (message.workflow_id != null && message.hasOwnProperty("workflow_id"))
                object.workflow_id = message.workflow_id;
            if (message.linked_workflow_id != null && message.hasOwnProperty("linked_workflow_id"))
                object.linked_workflow_id = message.linked_workflow_id;
            if (message.linked_workflow_name != null && message.hasOwnProperty("linked_workflow_name"))
                object.linked_workflow_name = message.linked_workflow_name;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this WorkflowLink to JSON.
         * @function toJSON
         * @memberof repository.WorkflowLink
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        WorkflowLink.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for WorkflowLink
         * @function getTypeUrl
         * @memberof repository.WorkflowLink
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WorkflowLink.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.WorkflowLink";
        };

        return WorkflowLink;
    })();

    repository.DependencyType = (function() {

        /**
         * Properties of a DependencyType.
         * @memberof repository
         * @interface IDependencyType
         * @property {string|null} [id] DependencyType id
         * @property {number|Long|null} [mtime] DependencyType mtime
         * @property {string|null} [name] DependencyType name
         * @property {boolean|null} [synced] DependencyType synced
         */

        /**
         * Constructs a new DependencyType.
         * @memberof repository
         * @classdesc Represents a DependencyType.
         * @implements IDependencyType
         * @constructor
         * @param {repository.IDependencyType=} [properties] Properties to set
         */
        function DependencyType(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * DependencyType id.
         * @member {string} id
         * @memberof repository.DependencyType
         * @instance
         */
        DependencyType.prototype.id = "";

        /**
         * DependencyType mtime.
         * @member {number|Long} mtime
         * @memberof repository.DependencyType
         * @instance
         */
        DependencyType.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * DependencyType name.
         * @member {string} name
         * @memberof repository.DependencyType
         * @instance
         */
        DependencyType.prototype.name = "";

        /**
         * DependencyType synced.
         * @member {boolean} synced
         * @memberof repository.DependencyType
         * @instance
         */
        DependencyType.prototype.synced = false;

        /**
         * Creates a new DependencyType instance using the specified properties.
         * @function create
         * @memberof repository.DependencyType
         * @static
         * @param {repository.IDependencyType=} [properties] Properties to set
         * @returns {repository.DependencyType} DependencyType instance
         */
        DependencyType.create = function create(properties) {
            return new DependencyType(properties);
        };

        /**
         * Encodes the specified DependencyType message. Does not implicitly {@link repository.DependencyType.verify|verify} messages.
         * @function encode
         * @memberof repository.DependencyType
         * @static
         * @param {repository.IDependencyType} message DependencyType message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DependencyType.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 4, wireType 0 =*/32).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified DependencyType message, length delimited. Does not implicitly {@link repository.DependencyType.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.DependencyType
         * @static
         * @param {repository.IDependencyType} message DependencyType message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        DependencyType.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a DependencyType message from the specified reader or buffer.
         * @function decode
         * @memberof repository.DependencyType
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.DependencyType} DependencyType
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DependencyType.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.DependencyType();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a DependencyType message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.DependencyType
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.DependencyType} DependencyType
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        DependencyType.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a DependencyType message.
         * @function verify
         * @memberof repository.DependencyType
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        DependencyType.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a DependencyType message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.DependencyType
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.DependencyType} DependencyType
         */
        DependencyType.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.DependencyType)
                return object;
            let message = new $root.repository.DependencyType();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a DependencyType message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.DependencyType
         * @static
         * @param {repository.DependencyType} message DependencyType
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        DependencyType.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this DependencyType to JSON.
         * @function toJSON
         * @memberof repository.DependencyType
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        DependencyType.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for DependencyType
         * @function getTypeUrl
         * @memberof repository.DependencyType
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        DependencyType.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.DependencyType";
        };

        return DependencyType;
    })();

    repository.Status = (function() {

        /**
         * Properties of a Status.
         * @memberof repository
         * @interface IStatus
         * @property {string|null} [id] Status id
         * @property {number|Long|null} [mtime] Status mtime
         * @property {string|null} [name] Status name
         * @property {string|null} [short_name] Status short_name
         * @property {string|null} [color] Status color
         * @property {boolean|null} [synced] Status synced
         */

        /**
         * Constructs a new Status.
         * @memberof repository
         * @classdesc Represents a Status.
         * @implements IStatus
         * @constructor
         * @param {repository.IStatus=} [properties] Properties to set
         */
        function Status(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Status id.
         * @member {string} id
         * @memberof repository.Status
         * @instance
         */
        Status.prototype.id = "";

        /**
         * Status mtime.
         * @member {number|Long} mtime
         * @memberof repository.Status
         * @instance
         */
        Status.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Status name.
         * @member {string} name
         * @memberof repository.Status
         * @instance
         */
        Status.prototype.name = "";

        /**
         * Status short_name.
         * @member {string} short_name
         * @memberof repository.Status
         * @instance
         */
        Status.prototype.short_name = "";

        /**
         * Status color.
         * @member {string} color
         * @memberof repository.Status
         * @instance
         */
        Status.prototype.color = "";

        /**
         * Status synced.
         * @member {boolean} synced
         * @memberof repository.Status
         * @instance
         */
        Status.prototype.synced = false;

        /**
         * Creates a new Status instance using the specified properties.
         * @function create
         * @memberof repository.Status
         * @static
         * @param {repository.IStatus=} [properties] Properties to set
         * @returns {repository.Status} Status instance
         */
        Status.create = function create(properties) {
            return new Status(properties);
        };

        /**
         * Encodes the specified Status message. Does not implicitly {@link repository.Status.verify|verify} messages.
         * @function encode
         * @memberof repository.Status
         * @static
         * @param {repository.IStatus} message Status message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Status.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.short_name != null && Object.hasOwnProperty.call(message, "short_name"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.short_name);
            if (message.color != null && Object.hasOwnProperty.call(message, "color"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.color);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified Status message, length delimited. Does not implicitly {@link repository.Status.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Status
         * @static
         * @param {repository.IStatus} message Status message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Status.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Status message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Status
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Status} Status
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Status.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Status();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.short_name = reader.string();
                        break;
                    }
                case 5: {
                        message.color = reader.string();
                        break;
                    }
                case 6: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Status message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Status
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Status} Status
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Status.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Status message.
         * @function verify
         * @memberof repository.Status
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Status.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.short_name != null && message.hasOwnProperty("short_name"))
                if (!$util.isString(message.short_name))
                    return "short_name: string expected";
            if (message.color != null && message.hasOwnProperty("color"))
                if (!$util.isString(message.color))
                    return "color: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a Status message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Status
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Status} Status
         */
        Status.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Status)
                return object;
            let message = new $root.repository.Status();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.short_name != null)
                message.short_name = String(object.short_name);
            if (object.color != null)
                message.color = String(object.color);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a Status message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Status
         * @static
         * @param {repository.Status} message Status
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Status.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.short_name = "";
                object.color = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.short_name != null && message.hasOwnProperty("short_name"))
                object.short_name = message.short_name;
            if (message.color != null && message.hasOwnProperty("color"))
                object.color = message.color;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this Status to JSON.
         * @function toJSON
         * @memberof repository.Status
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Status.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Status
         * @function getTypeUrl
         * @memberof repository.Status
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Status.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Status";
        };

        return Status;
    })();

    repository.Tag = (function() {

        /**
         * Properties of a Tag.
         * @memberof repository
         * @interface ITag
         * @property {string|null} [id] Tag id
         * @property {number|Long|null} [mtime] Tag mtime
         * @property {string|null} [name] Tag name
         * @property {boolean|null} [synced] Tag synced
         */

        /**
         * Constructs a new Tag.
         * @memberof repository
         * @classdesc Represents a Tag.
         * @implements ITag
         * @constructor
         * @param {repository.ITag=} [properties] Properties to set
         */
        function Tag(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Tag id.
         * @member {string} id
         * @memberof repository.Tag
         * @instance
         */
        Tag.prototype.id = "";

        /**
         * Tag mtime.
         * @member {number|Long} mtime
         * @memberof repository.Tag
         * @instance
         */
        Tag.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Tag name.
         * @member {string} name
         * @memberof repository.Tag
         * @instance
         */
        Tag.prototype.name = "";

        /**
         * Tag synced.
         * @member {boolean} synced
         * @memberof repository.Tag
         * @instance
         */
        Tag.prototype.synced = false;

        /**
         * Creates a new Tag instance using the specified properties.
         * @function create
         * @memberof repository.Tag
         * @static
         * @param {repository.ITag=} [properties] Properties to set
         * @returns {repository.Tag} Tag instance
         */
        Tag.create = function create(properties) {
            return new Tag(properties);
        };

        /**
         * Encodes the specified Tag message. Does not implicitly {@link repository.Tag.verify|verify} messages.
         * @function encode
         * @memberof repository.Tag
         * @static
         * @param {repository.ITag} message Tag message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Tag.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 4, wireType 0 =*/32).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified Tag message, length delimited. Does not implicitly {@link repository.Tag.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Tag
         * @static
         * @param {repository.ITag} message Tag message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Tag.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Tag message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Tag
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Tag} Tag
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Tag.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Tag();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Tag message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Tag
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Tag} Tag
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Tag.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Tag message.
         * @function verify
         * @memberof repository.Tag
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Tag.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a Tag message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Tag
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Tag} Tag
         */
        Tag.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Tag)
                return object;
            let message = new $root.repository.Tag();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a Tag message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Tag
         * @static
         * @param {repository.Tag} message Tag
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Tag.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this Tag to JSON.
         * @function toJSON
         * @memberof repository.Tag
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Tag.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Tag
         * @function getTypeUrl
         * @memberof repository.Tag
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Tag.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Tag";
        };

        return Tag;
    })();

    repository.TaskTag = (function() {

        /**
         * Properties of a TaskTag.
         * @memberof repository
         * @interface ITaskTag
         * @property {string|null} [id] TaskTag id
         * @property {number|Long|null} [mtime] TaskTag mtime
         * @property {string|null} [task_id] TaskTag task_id
         * @property {string|null} [tag_id] TaskTag tag_id
         * @property {boolean|null} [synced] TaskTag synced
         */

        /**
         * Constructs a new TaskTag.
         * @memberof repository
         * @classdesc Represents a TaskTag.
         * @implements ITaskTag
         * @constructor
         * @param {repository.ITaskTag=} [properties] Properties to set
         */
        function TaskTag(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * TaskTag id.
         * @member {string} id
         * @memberof repository.TaskTag
         * @instance
         */
        TaskTag.prototype.id = "";

        /**
         * TaskTag mtime.
         * @member {number|Long} mtime
         * @memberof repository.TaskTag
         * @instance
         */
        TaskTag.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * TaskTag task_id.
         * @member {string} task_id
         * @memberof repository.TaskTag
         * @instance
         */
        TaskTag.prototype.task_id = "";

        /**
         * TaskTag tag_id.
         * @member {string} tag_id
         * @memberof repository.TaskTag
         * @instance
         */
        TaskTag.prototype.tag_id = "";

        /**
         * TaskTag synced.
         * @member {boolean} synced
         * @memberof repository.TaskTag
         * @instance
         */
        TaskTag.prototype.synced = false;

        /**
         * Creates a new TaskTag instance using the specified properties.
         * @function create
         * @memberof repository.TaskTag
         * @static
         * @param {repository.ITaskTag=} [properties] Properties to set
         * @returns {repository.TaskTag} TaskTag instance
         */
        TaskTag.create = function create(properties) {
            return new TaskTag(properties);
        };

        /**
         * Encodes the specified TaskTag message. Does not implicitly {@link repository.TaskTag.verify|verify} messages.
         * @function encode
         * @memberof repository.TaskTag
         * @static
         * @param {repository.ITaskTag} message TaskTag message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TaskTag.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.task_id != null && Object.hasOwnProperty.call(message, "task_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.task_id);
            if (message.tag_id != null && Object.hasOwnProperty.call(message, "tag_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.tag_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 5, wireType 0 =*/40).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified TaskTag message, length delimited. Does not implicitly {@link repository.TaskTag.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.TaskTag
         * @static
         * @param {repository.ITaskTag} message TaskTag message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        TaskTag.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a TaskTag message from the specified reader or buffer.
         * @function decode
         * @memberof repository.TaskTag
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.TaskTag} TaskTag
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TaskTag.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.TaskTag();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.task_id = reader.string();
                        break;
                    }
                case 4: {
                        message.tag_id = reader.string();
                        break;
                    }
                case 5: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a TaskTag message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.TaskTag
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.TaskTag} TaskTag
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        TaskTag.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a TaskTag message.
         * @function verify
         * @memberof repository.TaskTag
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        TaskTag.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.task_id != null && message.hasOwnProperty("task_id"))
                if (!$util.isString(message.task_id))
                    return "task_id: string expected";
            if (message.tag_id != null && message.hasOwnProperty("tag_id"))
                if (!$util.isString(message.tag_id))
                    return "tag_id: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a TaskTag message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.TaskTag
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.TaskTag} TaskTag
         */
        TaskTag.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.TaskTag)
                return object;
            let message = new $root.repository.TaskTag();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.task_id != null)
                message.task_id = String(object.task_id);
            if (object.tag_id != null)
                message.tag_id = String(object.tag_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a TaskTag message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.TaskTag
         * @static
         * @param {repository.TaskTag} message TaskTag
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        TaskTag.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.task_id = "";
                object.tag_id = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.task_id != null && message.hasOwnProperty("task_id"))
                object.task_id = message.task_id;
            if (message.tag_id != null && message.hasOwnProperty("tag_id"))
                object.tag_id = message.tag_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this TaskTag to JSON.
         * @function toJSON
         * @memberof repository.TaskTag
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        TaskTag.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for TaskTag
         * @function getTypeUrl
         * @memberof repository.TaskTag
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        TaskTag.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.TaskTag";
        };

        return TaskTag;
    })();

    repository.Checkpoint = (function() {

        /**
         * Properties of a Checkpoint.
         * @memberof repository
         * @interface ICheckpoint
         * @property {string|null} [id] Checkpoint id
         * @property {number|Long|null} [mtime] Checkpoint mtime
         * @property {string|null} [created_at] Checkpoint created_at
         * @property {string|null} [task_id] Checkpoint task_id
         * @property {string|null} [xxhash_checksum] Checkpoint xxhash_checksum
         * @property {number|Long|null} [time_modified] Checkpoint time_modified
         * @property {number|Long|null} [file_size] Checkpoint file_size
         * @property {string|null} [comment] Checkpoint comment
         * @property {string|null} [chunks] Checkpoint chunks
         * @property {string|null} [author_uid] Checkpoint author_uid
         * @property {string|null} [preview_id] Checkpoint preview_id
         * @property {boolean|null} [trashed] Checkpoint trashed
         * @property {boolean|null} [synced] Checkpoint synced
         * @property {string|null} [group_id] Checkpoint group_id
         */

        /**
         * Constructs a new Checkpoint.
         * @memberof repository
         * @classdesc Represents a Checkpoint.
         * @implements ICheckpoint
         * @constructor
         * @param {repository.ICheckpoint=} [properties] Properties to set
         */
        function Checkpoint(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Checkpoint id.
         * @member {string} id
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.id = "";

        /**
         * Checkpoint mtime.
         * @member {number|Long} mtime
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Checkpoint created_at.
         * @member {string} created_at
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.created_at = "";

        /**
         * Checkpoint task_id.
         * @member {string} task_id
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.task_id = "";

        /**
         * Checkpoint xxhash_checksum.
         * @member {string} xxhash_checksum
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.xxhash_checksum = "";

        /**
         * Checkpoint time_modified.
         * @member {number|Long} time_modified
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.time_modified = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Checkpoint file_size.
         * @member {number|Long} file_size
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.file_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Checkpoint comment.
         * @member {string} comment
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.comment = "";

        /**
         * Checkpoint chunks.
         * @member {string} chunks
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.chunks = "";

        /**
         * Checkpoint author_uid.
         * @member {string} author_uid
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.author_uid = "";

        /**
         * Checkpoint preview_id.
         * @member {string} preview_id
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.preview_id = "";

        /**
         * Checkpoint trashed.
         * @member {boolean} trashed
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.trashed = false;

        /**
         * Checkpoint synced.
         * @member {boolean} synced
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.synced = false;

        /**
         * Checkpoint group_id.
         * @member {string} group_id
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.group_id = "";

        /**
         * Creates a new Checkpoint instance using the specified properties.
         * @function create
         * @memberof repository.Checkpoint
         * @static
         * @param {repository.ICheckpoint=} [properties] Properties to set
         * @returns {repository.Checkpoint} Checkpoint instance
         */
        Checkpoint.create = function create(properties) {
            return new Checkpoint(properties);
        };

        /**
         * Encodes the specified Checkpoint message. Does not implicitly {@link repository.Checkpoint.verify|verify} messages.
         * @function encode
         * @memberof repository.Checkpoint
         * @static
         * @param {repository.ICheckpoint} message Checkpoint message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Checkpoint.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.created_at);
            if (message.task_id != null && Object.hasOwnProperty.call(message, "task_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.task_id);
            if (message.xxhash_checksum != null && Object.hasOwnProperty.call(message, "xxhash_checksum"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.xxhash_checksum);
            if (message.time_modified != null && Object.hasOwnProperty.call(message, "time_modified"))
                writer.uint32(/* id 6, wireType 0 =*/48).int64(message.time_modified);
            if (message.file_size != null && Object.hasOwnProperty.call(message, "file_size"))
                writer.uint32(/* id 7, wireType 0 =*/56).int64(message.file_size);
            if (message.comment != null && Object.hasOwnProperty.call(message, "comment"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.comment);
            if (message.chunks != null && Object.hasOwnProperty.call(message, "chunks"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.chunks);
            if (message.author_uid != null && Object.hasOwnProperty.call(message, "author_uid"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.author_uid);
            if (message.preview_id != null && Object.hasOwnProperty.call(message, "preview_id"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.preview_id);
            if (message.trashed != null && Object.hasOwnProperty.call(message, "trashed"))
                writer.uint32(/* id 12, wireType 0 =*/96).bool(message.trashed);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 13, wireType 0 =*/104).bool(message.synced);
            if (message.group_id != null && Object.hasOwnProperty.call(message, "group_id"))
                writer.uint32(/* id 14, wireType 2 =*/114).string(message.group_id);
            return writer;
        };

        /**
         * Encodes the specified Checkpoint message, length delimited. Does not implicitly {@link repository.Checkpoint.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Checkpoint
         * @static
         * @param {repository.ICheckpoint} message Checkpoint message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Checkpoint.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Checkpoint message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Checkpoint
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Checkpoint} Checkpoint
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Checkpoint.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Checkpoint();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.created_at = reader.string();
                        break;
                    }
                case 4: {
                        message.task_id = reader.string();
                        break;
                    }
                case 5: {
                        message.xxhash_checksum = reader.string();
                        break;
                    }
                case 6: {
                        message.time_modified = reader.int64();
                        break;
                    }
                case 7: {
                        message.file_size = reader.int64();
                        break;
                    }
                case 8: {
                        message.comment = reader.string();
                        break;
                    }
                case 9: {
                        message.chunks = reader.string();
                        break;
                    }
                case 10: {
                        message.author_uid = reader.string();
                        break;
                    }
                case 11: {
                        message.preview_id = reader.string();
                        break;
                    }
                case 12: {
                        message.trashed = reader.bool();
                        break;
                    }
                case 13: {
                        message.synced = reader.bool();
                        break;
                    }
                case 14: {
                        message.group_id = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Checkpoint message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Checkpoint
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Checkpoint} Checkpoint
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Checkpoint.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Checkpoint message.
         * @function verify
         * @memberof repository.Checkpoint
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Checkpoint.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.created_at != null && message.hasOwnProperty("created_at"))
                if (!$util.isString(message.created_at))
                    return "created_at: string expected";
            if (message.task_id != null && message.hasOwnProperty("task_id"))
                if (!$util.isString(message.task_id))
                    return "task_id: string expected";
            if (message.xxhash_checksum != null && message.hasOwnProperty("xxhash_checksum"))
                if (!$util.isString(message.xxhash_checksum))
                    return "xxhash_checksum: string expected";
            if (message.time_modified != null && message.hasOwnProperty("time_modified"))
                if (!$util.isInteger(message.time_modified) && !(message.time_modified && $util.isInteger(message.time_modified.low) && $util.isInteger(message.time_modified.high)))
                    return "time_modified: integer|Long expected";
            if (message.file_size != null && message.hasOwnProperty("file_size"))
                if (!$util.isInteger(message.file_size) && !(message.file_size && $util.isInteger(message.file_size.low) && $util.isInteger(message.file_size.high)))
                    return "file_size: integer|Long expected";
            if (message.comment != null && message.hasOwnProperty("comment"))
                if (!$util.isString(message.comment))
                    return "comment: string expected";
            if (message.chunks != null && message.hasOwnProperty("chunks"))
                if (!$util.isString(message.chunks))
                    return "chunks: string expected";
            if (message.author_uid != null && message.hasOwnProperty("author_uid"))
                if (!$util.isString(message.author_uid))
                    return "author_uid: string expected";
            if (message.preview_id != null && message.hasOwnProperty("preview_id"))
                if (!$util.isString(message.preview_id))
                    return "preview_id: string expected";
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                if (typeof message.trashed !== "boolean")
                    return "trashed: boolean expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            if (message.group_id != null && message.hasOwnProperty("group_id"))
                if (!$util.isString(message.group_id))
                    return "group_id: string expected";
            return null;
        };

        /**
         * Creates a Checkpoint message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Checkpoint
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Checkpoint} Checkpoint
         */
        Checkpoint.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Checkpoint)
                return object;
            let message = new $root.repository.Checkpoint();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.created_at != null)
                message.created_at = String(object.created_at);
            if (object.task_id != null)
                message.task_id = String(object.task_id);
            if (object.xxhash_checksum != null)
                message.xxhash_checksum = String(object.xxhash_checksum);
            if (object.time_modified != null)
                if ($util.Long)
                    (message.time_modified = $util.Long.fromValue(object.time_modified)).unsigned = false;
                else if (typeof object.time_modified === "string")
                    message.time_modified = parseInt(object.time_modified, 10);
                else if (typeof object.time_modified === "number")
                    message.time_modified = object.time_modified;
                else if (typeof object.time_modified === "object")
                    message.time_modified = new $util.LongBits(object.time_modified.low >>> 0, object.time_modified.high >>> 0).toNumber();
            if (object.file_size != null)
                if ($util.Long)
                    (message.file_size = $util.Long.fromValue(object.file_size)).unsigned = false;
                else if (typeof object.file_size === "string")
                    message.file_size = parseInt(object.file_size, 10);
                else if (typeof object.file_size === "number")
                    message.file_size = object.file_size;
                else if (typeof object.file_size === "object")
                    message.file_size = new $util.LongBits(object.file_size.low >>> 0, object.file_size.high >>> 0).toNumber();
            if (object.comment != null)
                message.comment = String(object.comment);
            if (object.chunks != null)
                message.chunks = String(object.chunks);
            if (object.author_uid != null)
                message.author_uid = String(object.author_uid);
            if (object.preview_id != null)
                message.preview_id = String(object.preview_id);
            if (object.trashed != null)
                message.trashed = Boolean(object.trashed);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            if (object.group_id != null)
                message.group_id = String(object.group_id);
            return message;
        };

        /**
         * Creates a plain object from a Checkpoint message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Checkpoint
         * @static
         * @param {repository.Checkpoint} message Checkpoint
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Checkpoint.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.created_at = "";
                object.task_id = "";
                object.xxhash_checksum = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.time_modified = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.time_modified = options.longs === String ? "0" : 0;
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.file_size = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.file_size = options.longs === String ? "0" : 0;
                object.comment = "";
                object.chunks = "";
                object.author_uid = "";
                object.preview_id = "";
                object.trashed = false;
                object.synced = false;
                object.group_id = "";
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.created_at != null && message.hasOwnProperty("created_at"))
                object.created_at = message.created_at;
            if (message.task_id != null && message.hasOwnProperty("task_id"))
                object.task_id = message.task_id;
            if (message.xxhash_checksum != null && message.hasOwnProperty("xxhash_checksum"))
                object.xxhash_checksum = message.xxhash_checksum;
            if (message.time_modified != null && message.hasOwnProperty("time_modified"))
                if (typeof message.time_modified === "number")
                    object.time_modified = options.longs === String ? String(message.time_modified) : message.time_modified;
                else
                    object.time_modified = options.longs === String ? $util.Long.prototype.toString.call(message.time_modified) : options.longs === Number ? new $util.LongBits(message.time_modified.low >>> 0, message.time_modified.high >>> 0).toNumber() : message.time_modified;
            if (message.file_size != null && message.hasOwnProperty("file_size"))
                if (typeof message.file_size === "number")
                    object.file_size = options.longs === String ? String(message.file_size) : message.file_size;
                else
                    object.file_size = options.longs === String ? $util.Long.prototype.toString.call(message.file_size) : options.longs === Number ? new $util.LongBits(message.file_size.low >>> 0, message.file_size.high >>> 0).toNumber() : message.file_size;
            if (message.comment != null && message.hasOwnProperty("comment"))
                object.comment = message.comment;
            if (message.chunks != null && message.hasOwnProperty("chunks"))
                object.chunks = message.chunks;
            if (message.author_uid != null && message.hasOwnProperty("author_uid"))
                object.author_uid = message.author_uid;
            if (message.preview_id != null && message.hasOwnProperty("preview_id"))
                object.preview_id = message.preview_id;
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                object.trashed = message.trashed;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            if (message.group_id != null && message.hasOwnProperty("group_id"))
                object.group_id = message.group_id;
            return object;
        };

        /**
         * Converts this Checkpoint to JSON.
         * @function toJSON
         * @memberof repository.Checkpoint
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Checkpoint.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Checkpoint
         * @function getTypeUrl
         * @memberof repository.Checkpoint
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Checkpoint.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Checkpoint";
        };

        return Checkpoint;
    })();

    repository.Role = (function() {

        /**
         * Properties of a Role.
         * @memberof repository
         * @interface IRole
         * @property {string|null} [id] Role id
         * @property {number|Long|null} [mtime] Role mtime
         * @property {string|null} [name] Role name
         * @property {boolean|null} [synced] Role synced
         * @property {boolean|null} [view_entity] Role view_entity
         * @property {boolean|null} [create_entity] Role create_entity
         * @property {boolean|null} [update_entity] Role update_entity
         * @property {boolean|null} [delete_entity] Role delete_entity
         * @property {boolean|null} [view_task] Role view_task
         * @property {boolean|null} [create_task] Role create_task
         * @property {boolean|null} [update_task] Role update_task
         * @property {boolean|null} [delete_task] Role delete_task
         * @property {boolean|null} [view_template] Role view_template
         * @property {boolean|null} [create_template] Role create_template
         * @property {boolean|null} [update_template] Role update_template
         * @property {boolean|null} [delete_template] Role delete_template
         * @property {boolean|null} [view_checkpoint] Role view_checkpoint
         * @property {boolean|null} [create_checkpoint] Role create_checkpoint
         * @property {boolean|null} [delete_checkpoint] Role delete_checkpoint
         * @property {boolean|null} [pull_chunk] Role pull_chunk
         * @property {boolean|null} [assign_task] Role assign_task
         * @property {boolean|null} [unassign_task] Role unassign_task
         * @property {boolean|null} [add_user] Role add_user
         * @property {boolean|null} [remove_user] Role remove_user
         * @property {boolean|null} [change_role] Role change_role
         * @property {boolean|null} [change_status] Role change_status
         * @property {boolean|null} [set_done_task] Role set_done_task
         * @property {boolean|null} [set_retake_task] Role set_retake_task
         * @property {boolean|null} [view_done_task] Role view_done_task
         * @property {boolean|null} [manage_dependencies] Role manage_dependencies
         */

        /**
         * Constructs a new Role.
         * @memberof repository
         * @classdesc Represents a Role.
         * @implements IRole
         * @constructor
         * @param {repository.IRole=} [properties] Properties to set
         */
        function Role(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Role id.
         * @member {string} id
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.id = "";

        /**
         * Role mtime.
         * @member {number|Long} mtime
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Role name.
         * @member {string} name
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.name = "";

        /**
         * Role synced.
         * @member {boolean} synced
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.synced = false;

        /**
         * Role view_entity.
         * @member {boolean} view_entity
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.view_entity = false;

        /**
         * Role create_entity.
         * @member {boolean} create_entity
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.create_entity = false;

        /**
         * Role update_entity.
         * @member {boolean} update_entity
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.update_entity = false;

        /**
         * Role delete_entity.
         * @member {boolean} delete_entity
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.delete_entity = false;

        /**
         * Role view_task.
         * @member {boolean} view_task
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.view_task = false;

        /**
         * Role create_task.
         * @member {boolean} create_task
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.create_task = false;

        /**
         * Role update_task.
         * @member {boolean} update_task
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.update_task = false;

        /**
         * Role delete_task.
         * @member {boolean} delete_task
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.delete_task = false;

        /**
         * Role view_template.
         * @member {boolean} view_template
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.view_template = false;

        /**
         * Role create_template.
         * @member {boolean} create_template
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.create_template = false;

        /**
         * Role update_template.
         * @member {boolean} update_template
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.update_template = false;

        /**
         * Role delete_template.
         * @member {boolean} delete_template
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.delete_template = false;

        /**
         * Role view_checkpoint.
         * @member {boolean} view_checkpoint
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.view_checkpoint = false;

        /**
         * Role create_checkpoint.
         * @member {boolean} create_checkpoint
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.create_checkpoint = false;

        /**
         * Role delete_checkpoint.
         * @member {boolean} delete_checkpoint
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.delete_checkpoint = false;

        /**
         * Role pull_chunk.
         * @member {boolean} pull_chunk
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.pull_chunk = false;

        /**
         * Role assign_task.
         * @member {boolean} assign_task
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.assign_task = false;

        /**
         * Role unassign_task.
         * @member {boolean} unassign_task
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.unassign_task = false;

        /**
         * Role add_user.
         * @member {boolean} add_user
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.add_user = false;

        /**
         * Role remove_user.
         * @member {boolean} remove_user
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.remove_user = false;

        /**
         * Role change_role.
         * @member {boolean} change_role
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.change_role = false;

        /**
         * Role change_status.
         * @member {boolean} change_status
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.change_status = false;

        /**
         * Role set_done_task.
         * @member {boolean} set_done_task
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.set_done_task = false;

        /**
         * Role set_retake_task.
         * @member {boolean} set_retake_task
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.set_retake_task = false;

        /**
         * Role view_done_task.
         * @member {boolean} view_done_task
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.view_done_task = false;

        /**
         * Role manage_dependencies.
         * @member {boolean} manage_dependencies
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.manage_dependencies = false;

        /**
         * Creates a new Role instance using the specified properties.
         * @function create
         * @memberof repository.Role
         * @static
         * @param {repository.IRole=} [properties] Properties to set
         * @returns {repository.Role} Role instance
         */
        Role.create = function create(properties) {
            return new Role(properties);
        };

        /**
         * Encodes the specified Role message. Does not implicitly {@link repository.Role.verify|verify} messages.
         * @function encode
         * @memberof repository.Role
         * @static
         * @param {repository.IRole} message Role message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Role.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 4, wireType 0 =*/32).bool(message.synced);
            if (message.view_entity != null && Object.hasOwnProperty.call(message, "view_entity"))
                writer.uint32(/* id 5, wireType 0 =*/40).bool(message.view_entity);
            if (message.create_entity != null && Object.hasOwnProperty.call(message, "create_entity"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.create_entity);
            if (message.update_entity != null && Object.hasOwnProperty.call(message, "update_entity"))
                writer.uint32(/* id 7, wireType 0 =*/56).bool(message.update_entity);
            if (message.delete_entity != null && Object.hasOwnProperty.call(message, "delete_entity"))
                writer.uint32(/* id 8, wireType 0 =*/64).bool(message.delete_entity);
            if (message.view_task != null && Object.hasOwnProperty.call(message, "view_task"))
                writer.uint32(/* id 9, wireType 0 =*/72).bool(message.view_task);
            if (message.create_task != null && Object.hasOwnProperty.call(message, "create_task"))
                writer.uint32(/* id 10, wireType 0 =*/80).bool(message.create_task);
            if (message.update_task != null && Object.hasOwnProperty.call(message, "update_task"))
                writer.uint32(/* id 11, wireType 0 =*/88).bool(message.update_task);
            if (message.delete_task != null && Object.hasOwnProperty.call(message, "delete_task"))
                writer.uint32(/* id 12, wireType 0 =*/96).bool(message.delete_task);
            if (message.view_template != null && Object.hasOwnProperty.call(message, "view_template"))
                writer.uint32(/* id 13, wireType 0 =*/104).bool(message.view_template);
            if (message.create_template != null && Object.hasOwnProperty.call(message, "create_template"))
                writer.uint32(/* id 14, wireType 0 =*/112).bool(message.create_template);
            if (message.update_template != null && Object.hasOwnProperty.call(message, "update_template"))
                writer.uint32(/* id 15, wireType 0 =*/120).bool(message.update_template);
            if (message.delete_template != null && Object.hasOwnProperty.call(message, "delete_template"))
                writer.uint32(/* id 16, wireType 0 =*/128).bool(message.delete_template);
            if (message.view_checkpoint != null && Object.hasOwnProperty.call(message, "view_checkpoint"))
                writer.uint32(/* id 17, wireType 0 =*/136).bool(message.view_checkpoint);
            if (message.create_checkpoint != null && Object.hasOwnProperty.call(message, "create_checkpoint"))
                writer.uint32(/* id 18, wireType 0 =*/144).bool(message.create_checkpoint);
            if (message.delete_checkpoint != null && Object.hasOwnProperty.call(message, "delete_checkpoint"))
                writer.uint32(/* id 19, wireType 0 =*/152).bool(message.delete_checkpoint);
            if (message.pull_chunk != null && Object.hasOwnProperty.call(message, "pull_chunk"))
                writer.uint32(/* id 20, wireType 0 =*/160).bool(message.pull_chunk);
            if (message.assign_task != null && Object.hasOwnProperty.call(message, "assign_task"))
                writer.uint32(/* id 21, wireType 0 =*/168).bool(message.assign_task);
            if (message.unassign_task != null && Object.hasOwnProperty.call(message, "unassign_task"))
                writer.uint32(/* id 22, wireType 0 =*/176).bool(message.unassign_task);
            if (message.add_user != null && Object.hasOwnProperty.call(message, "add_user"))
                writer.uint32(/* id 23, wireType 0 =*/184).bool(message.add_user);
            if (message.remove_user != null && Object.hasOwnProperty.call(message, "remove_user"))
                writer.uint32(/* id 24, wireType 0 =*/192).bool(message.remove_user);
            if (message.change_role != null && Object.hasOwnProperty.call(message, "change_role"))
                writer.uint32(/* id 25, wireType 0 =*/200).bool(message.change_role);
            if (message.change_status != null && Object.hasOwnProperty.call(message, "change_status"))
                writer.uint32(/* id 26, wireType 0 =*/208).bool(message.change_status);
            if (message.set_done_task != null && Object.hasOwnProperty.call(message, "set_done_task"))
                writer.uint32(/* id 27, wireType 0 =*/216).bool(message.set_done_task);
            if (message.set_retake_task != null && Object.hasOwnProperty.call(message, "set_retake_task"))
                writer.uint32(/* id 28, wireType 0 =*/224).bool(message.set_retake_task);
            if (message.view_done_task != null && Object.hasOwnProperty.call(message, "view_done_task"))
                writer.uint32(/* id 29, wireType 0 =*/232).bool(message.view_done_task);
            if (message.manage_dependencies != null && Object.hasOwnProperty.call(message, "manage_dependencies"))
                writer.uint32(/* id 30, wireType 0 =*/240).bool(message.manage_dependencies);
            return writer;
        };

        /**
         * Encodes the specified Role message, length delimited. Does not implicitly {@link repository.Role.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Role
         * @static
         * @param {repository.IRole} message Role message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Role.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Role message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Role
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Role} Role
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Role.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Role();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.synced = reader.bool();
                        break;
                    }
                case 5: {
                        message.view_entity = reader.bool();
                        break;
                    }
                case 6: {
                        message.create_entity = reader.bool();
                        break;
                    }
                case 7: {
                        message.update_entity = reader.bool();
                        break;
                    }
                case 8: {
                        message.delete_entity = reader.bool();
                        break;
                    }
                case 9: {
                        message.view_task = reader.bool();
                        break;
                    }
                case 10: {
                        message.create_task = reader.bool();
                        break;
                    }
                case 11: {
                        message.update_task = reader.bool();
                        break;
                    }
                case 12: {
                        message.delete_task = reader.bool();
                        break;
                    }
                case 13: {
                        message.view_template = reader.bool();
                        break;
                    }
                case 14: {
                        message.create_template = reader.bool();
                        break;
                    }
                case 15: {
                        message.update_template = reader.bool();
                        break;
                    }
                case 16: {
                        message.delete_template = reader.bool();
                        break;
                    }
                case 17: {
                        message.view_checkpoint = reader.bool();
                        break;
                    }
                case 18: {
                        message.create_checkpoint = reader.bool();
                        break;
                    }
                case 19: {
                        message.delete_checkpoint = reader.bool();
                        break;
                    }
                case 20: {
                        message.pull_chunk = reader.bool();
                        break;
                    }
                case 21: {
                        message.assign_task = reader.bool();
                        break;
                    }
                case 22: {
                        message.unassign_task = reader.bool();
                        break;
                    }
                case 23: {
                        message.add_user = reader.bool();
                        break;
                    }
                case 24: {
                        message.remove_user = reader.bool();
                        break;
                    }
                case 25: {
                        message.change_role = reader.bool();
                        break;
                    }
                case 26: {
                        message.change_status = reader.bool();
                        break;
                    }
                case 27: {
                        message.set_done_task = reader.bool();
                        break;
                    }
                case 28: {
                        message.set_retake_task = reader.bool();
                        break;
                    }
                case 29: {
                        message.view_done_task = reader.bool();
                        break;
                    }
                case 30: {
                        message.manage_dependencies = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Role message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Role
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Role} Role
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Role.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Role message.
         * @function verify
         * @memberof repository.Role
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Role.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            if (message.view_entity != null && message.hasOwnProperty("view_entity"))
                if (typeof message.view_entity !== "boolean")
                    return "view_entity: boolean expected";
            if (message.create_entity != null && message.hasOwnProperty("create_entity"))
                if (typeof message.create_entity !== "boolean")
                    return "create_entity: boolean expected";
            if (message.update_entity != null && message.hasOwnProperty("update_entity"))
                if (typeof message.update_entity !== "boolean")
                    return "update_entity: boolean expected";
            if (message.delete_entity != null && message.hasOwnProperty("delete_entity"))
                if (typeof message.delete_entity !== "boolean")
                    return "delete_entity: boolean expected";
            if (message.view_task != null && message.hasOwnProperty("view_task"))
                if (typeof message.view_task !== "boolean")
                    return "view_task: boolean expected";
            if (message.create_task != null && message.hasOwnProperty("create_task"))
                if (typeof message.create_task !== "boolean")
                    return "create_task: boolean expected";
            if (message.update_task != null && message.hasOwnProperty("update_task"))
                if (typeof message.update_task !== "boolean")
                    return "update_task: boolean expected";
            if (message.delete_task != null && message.hasOwnProperty("delete_task"))
                if (typeof message.delete_task !== "boolean")
                    return "delete_task: boolean expected";
            if (message.view_template != null && message.hasOwnProperty("view_template"))
                if (typeof message.view_template !== "boolean")
                    return "view_template: boolean expected";
            if (message.create_template != null && message.hasOwnProperty("create_template"))
                if (typeof message.create_template !== "boolean")
                    return "create_template: boolean expected";
            if (message.update_template != null && message.hasOwnProperty("update_template"))
                if (typeof message.update_template !== "boolean")
                    return "update_template: boolean expected";
            if (message.delete_template != null && message.hasOwnProperty("delete_template"))
                if (typeof message.delete_template !== "boolean")
                    return "delete_template: boolean expected";
            if (message.view_checkpoint != null && message.hasOwnProperty("view_checkpoint"))
                if (typeof message.view_checkpoint !== "boolean")
                    return "view_checkpoint: boolean expected";
            if (message.create_checkpoint != null && message.hasOwnProperty("create_checkpoint"))
                if (typeof message.create_checkpoint !== "boolean")
                    return "create_checkpoint: boolean expected";
            if (message.delete_checkpoint != null && message.hasOwnProperty("delete_checkpoint"))
                if (typeof message.delete_checkpoint !== "boolean")
                    return "delete_checkpoint: boolean expected";
            if (message.pull_chunk != null && message.hasOwnProperty("pull_chunk"))
                if (typeof message.pull_chunk !== "boolean")
                    return "pull_chunk: boolean expected";
            if (message.assign_task != null && message.hasOwnProperty("assign_task"))
                if (typeof message.assign_task !== "boolean")
                    return "assign_task: boolean expected";
            if (message.unassign_task != null && message.hasOwnProperty("unassign_task"))
                if (typeof message.unassign_task !== "boolean")
                    return "unassign_task: boolean expected";
            if (message.add_user != null && message.hasOwnProperty("add_user"))
                if (typeof message.add_user !== "boolean")
                    return "add_user: boolean expected";
            if (message.remove_user != null && message.hasOwnProperty("remove_user"))
                if (typeof message.remove_user !== "boolean")
                    return "remove_user: boolean expected";
            if (message.change_role != null && message.hasOwnProperty("change_role"))
                if (typeof message.change_role !== "boolean")
                    return "change_role: boolean expected";
            if (message.change_status != null && message.hasOwnProperty("change_status"))
                if (typeof message.change_status !== "boolean")
                    return "change_status: boolean expected";
            if (message.set_done_task != null && message.hasOwnProperty("set_done_task"))
                if (typeof message.set_done_task !== "boolean")
                    return "set_done_task: boolean expected";
            if (message.set_retake_task != null && message.hasOwnProperty("set_retake_task"))
                if (typeof message.set_retake_task !== "boolean")
                    return "set_retake_task: boolean expected";
            if (message.view_done_task != null && message.hasOwnProperty("view_done_task"))
                if (typeof message.view_done_task !== "boolean")
                    return "view_done_task: boolean expected";
            if (message.manage_dependencies != null && message.hasOwnProperty("manage_dependencies"))
                if (typeof message.manage_dependencies !== "boolean")
                    return "manage_dependencies: boolean expected";
            return null;
        };

        /**
         * Creates a Role message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Role
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Role} Role
         */
        Role.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Role)
                return object;
            let message = new $root.repository.Role();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            if (object.view_entity != null)
                message.view_entity = Boolean(object.view_entity);
            if (object.create_entity != null)
                message.create_entity = Boolean(object.create_entity);
            if (object.update_entity != null)
                message.update_entity = Boolean(object.update_entity);
            if (object.delete_entity != null)
                message.delete_entity = Boolean(object.delete_entity);
            if (object.view_task != null)
                message.view_task = Boolean(object.view_task);
            if (object.create_task != null)
                message.create_task = Boolean(object.create_task);
            if (object.update_task != null)
                message.update_task = Boolean(object.update_task);
            if (object.delete_task != null)
                message.delete_task = Boolean(object.delete_task);
            if (object.view_template != null)
                message.view_template = Boolean(object.view_template);
            if (object.create_template != null)
                message.create_template = Boolean(object.create_template);
            if (object.update_template != null)
                message.update_template = Boolean(object.update_template);
            if (object.delete_template != null)
                message.delete_template = Boolean(object.delete_template);
            if (object.view_checkpoint != null)
                message.view_checkpoint = Boolean(object.view_checkpoint);
            if (object.create_checkpoint != null)
                message.create_checkpoint = Boolean(object.create_checkpoint);
            if (object.delete_checkpoint != null)
                message.delete_checkpoint = Boolean(object.delete_checkpoint);
            if (object.pull_chunk != null)
                message.pull_chunk = Boolean(object.pull_chunk);
            if (object.assign_task != null)
                message.assign_task = Boolean(object.assign_task);
            if (object.unassign_task != null)
                message.unassign_task = Boolean(object.unassign_task);
            if (object.add_user != null)
                message.add_user = Boolean(object.add_user);
            if (object.remove_user != null)
                message.remove_user = Boolean(object.remove_user);
            if (object.change_role != null)
                message.change_role = Boolean(object.change_role);
            if (object.change_status != null)
                message.change_status = Boolean(object.change_status);
            if (object.set_done_task != null)
                message.set_done_task = Boolean(object.set_done_task);
            if (object.set_retake_task != null)
                message.set_retake_task = Boolean(object.set_retake_task);
            if (object.view_done_task != null)
                message.view_done_task = Boolean(object.view_done_task);
            if (object.manage_dependencies != null)
                message.manage_dependencies = Boolean(object.manage_dependencies);
            return message;
        };

        /**
         * Creates a plain object from a Role message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Role
         * @static
         * @param {repository.Role} message Role
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Role.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.synced = false;
                object.view_entity = false;
                object.create_entity = false;
                object.update_entity = false;
                object.delete_entity = false;
                object.view_task = false;
                object.create_task = false;
                object.update_task = false;
                object.delete_task = false;
                object.view_template = false;
                object.create_template = false;
                object.update_template = false;
                object.delete_template = false;
                object.view_checkpoint = false;
                object.create_checkpoint = false;
                object.delete_checkpoint = false;
                object.pull_chunk = false;
                object.assign_task = false;
                object.unassign_task = false;
                object.add_user = false;
                object.remove_user = false;
                object.change_role = false;
                object.change_status = false;
                object.set_done_task = false;
                object.set_retake_task = false;
                object.view_done_task = false;
                object.manage_dependencies = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            if (message.view_entity != null && message.hasOwnProperty("view_entity"))
                object.view_entity = message.view_entity;
            if (message.create_entity != null && message.hasOwnProperty("create_entity"))
                object.create_entity = message.create_entity;
            if (message.update_entity != null && message.hasOwnProperty("update_entity"))
                object.update_entity = message.update_entity;
            if (message.delete_entity != null && message.hasOwnProperty("delete_entity"))
                object.delete_entity = message.delete_entity;
            if (message.view_task != null && message.hasOwnProperty("view_task"))
                object.view_task = message.view_task;
            if (message.create_task != null && message.hasOwnProperty("create_task"))
                object.create_task = message.create_task;
            if (message.update_task != null && message.hasOwnProperty("update_task"))
                object.update_task = message.update_task;
            if (message.delete_task != null && message.hasOwnProperty("delete_task"))
                object.delete_task = message.delete_task;
            if (message.view_template != null && message.hasOwnProperty("view_template"))
                object.view_template = message.view_template;
            if (message.create_template != null && message.hasOwnProperty("create_template"))
                object.create_template = message.create_template;
            if (message.update_template != null && message.hasOwnProperty("update_template"))
                object.update_template = message.update_template;
            if (message.delete_template != null && message.hasOwnProperty("delete_template"))
                object.delete_template = message.delete_template;
            if (message.view_checkpoint != null && message.hasOwnProperty("view_checkpoint"))
                object.view_checkpoint = message.view_checkpoint;
            if (message.create_checkpoint != null && message.hasOwnProperty("create_checkpoint"))
                object.create_checkpoint = message.create_checkpoint;
            if (message.delete_checkpoint != null && message.hasOwnProperty("delete_checkpoint"))
                object.delete_checkpoint = message.delete_checkpoint;
            if (message.pull_chunk != null && message.hasOwnProperty("pull_chunk"))
                object.pull_chunk = message.pull_chunk;
            if (message.assign_task != null && message.hasOwnProperty("assign_task"))
                object.assign_task = message.assign_task;
            if (message.unassign_task != null && message.hasOwnProperty("unassign_task"))
                object.unassign_task = message.unassign_task;
            if (message.add_user != null && message.hasOwnProperty("add_user"))
                object.add_user = message.add_user;
            if (message.remove_user != null && message.hasOwnProperty("remove_user"))
                object.remove_user = message.remove_user;
            if (message.change_role != null && message.hasOwnProperty("change_role"))
                object.change_role = message.change_role;
            if (message.change_status != null && message.hasOwnProperty("change_status"))
                object.change_status = message.change_status;
            if (message.set_done_task != null && message.hasOwnProperty("set_done_task"))
                object.set_done_task = message.set_done_task;
            if (message.set_retake_task != null && message.hasOwnProperty("set_retake_task"))
                object.set_retake_task = message.set_retake_task;
            if (message.view_done_task != null && message.hasOwnProperty("view_done_task"))
                object.view_done_task = message.view_done_task;
            if (message.manage_dependencies != null && message.hasOwnProperty("manage_dependencies"))
                object.manage_dependencies = message.manage_dependencies;
            return object;
        };

        /**
         * Converts this Role to JSON.
         * @function toJSON
         * @memberof repository.Role
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Role.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Role
         * @function getTypeUrl
         * @memberof repository.Role
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Role.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Role";
        };

        return Role;
    })();

    repository.UserRole = (function() {

        /**
         * Properties of a UserRole.
         * @memberof repository
         * @interface IUserRole
         * @property {string|null} [id] UserRole id
         * @property {number|Long|null} [mtime] UserRole mtime
         * @property {string|null} [user_id] UserRole user_id
         * @property {string|null} [role_id] UserRole role_id
         * @property {boolean|null} [synced] UserRole synced
         */

        /**
         * Constructs a new UserRole.
         * @memberof repository
         * @classdesc Represents a UserRole.
         * @implements IUserRole
         * @constructor
         * @param {repository.IUserRole=} [properties] Properties to set
         */
        function UserRole(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * UserRole id.
         * @member {string} id
         * @memberof repository.UserRole
         * @instance
         */
        UserRole.prototype.id = "";

        /**
         * UserRole mtime.
         * @member {number|Long} mtime
         * @memberof repository.UserRole
         * @instance
         */
        UserRole.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * UserRole user_id.
         * @member {string} user_id
         * @memberof repository.UserRole
         * @instance
         */
        UserRole.prototype.user_id = "";

        /**
         * UserRole role_id.
         * @member {string} role_id
         * @memberof repository.UserRole
         * @instance
         */
        UserRole.prototype.role_id = "";

        /**
         * UserRole synced.
         * @member {boolean} synced
         * @memberof repository.UserRole
         * @instance
         */
        UserRole.prototype.synced = false;

        /**
         * Creates a new UserRole instance using the specified properties.
         * @function create
         * @memberof repository.UserRole
         * @static
         * @param {repository.IUserRole=} [properties] Properties to set
         * @returns {repository.UserRole} UserRole instance
         */
        UserRole.create = function create(properties) {
            return new UserRole(properties);
        };

        /**
         * Encodes the specified UserRole message. Does not implicitly {@link repository.UserRole.verify|verify} messages.
         * @function encode
         * @memberof repository.UserRole
         * @static
         * @param {repository.IUserRole} message UserRole message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        UserRole.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.user_id != null && Object.hasOwnProperty.call(message, "user_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.user_id);
            if (message.role_id != null && Object.hasOwnProperty.call(message, "role_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.role_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 5, wireType 0 =*/40).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified UserRole message, length delimited. Does not implicitly {@link repository.UserRole.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.UserRole
         * @static
         * @param {repository.IUserRole} message UserRole message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        UserRole.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a UserRole message from the specified reader or buffer.
         * @function decode
         * @memberof repository.UserRole
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.UserRole} UserRole
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UserRole.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.UserRole();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.user_id = reader.string();
                        break;
                    }
                case 4: {
                        message.role_id = reader.string();
                        break;
                    }
                case 5: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a UserRole message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.UserRole
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.UserRole} UserRole
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        UserRole.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a UserRole message.
         * @function verify
         * @memberof repository.UserRole
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        UserRole.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.user_id != null && message.hasOwnProperty("user_id"))
                if (!$util.isString(message.user_id))
                    return "user_id: string expected";
            if (message.role_id != null && message.hasOwnProperty("role_id"))
                if (!$util.isString(message.role_id))
                    return "role_id: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a UserRole message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.UserRole
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.UserRole} UserRole
         */
        UserRole.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.UserRole)
                return object;
            let message = new $root.repository.UserRole();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.user_id != null)
                message.user_id = String(object.user_id);
            if (object.role_id != null)
                message.role_id = String(object.role_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a UserRole message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.UserRole
         * @static
         * @param {repository.UserRole} message UserRole
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        UserRole.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.user_id = "";
                object.role_id = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.user_id != null && message.hasOwnProperty("user_id"))
                object.user_id = message.user_id;
            if (message.role_id != null && message.hasOwnProperty("role_id"))
                object.role_id = message.role_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this UserRole to JSON.
         * @function toJSON
         * @memberof repository.UserRole
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        UserRole.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for UserRole
         * @function getTypeUrl
         * @memberof repository.UserRole
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        UserRole.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.UserRole";
        };

        return UserRole;
    })();

    repository.Template = (function() {

        /**
         * Properties of a Template.
         * @memberof repository
         * @interface ITemplate
         * @property {string|null} [id] Template id
         * @property {number|Long|null} [mtime] Template mtime
         * @property {string|null} [name] Template name
         * @property {string|null} [extension] Template extension
         * @property {string|null} [chunks] Template chunks
         * @property {string|null} [xxhash_checksum] Template xxhash_checksum
         * @property {number|Long|null} [file_size] Template file_size
         * @property {boolean|null} [trashed] Template trashed
         * @property {boolean|null} [synced] Template synced
         */

        /**
         * Constructs a new Template.
         * @memberof repository
         * @classdesc Represents a Template.
         * @implements ITemplate
         * @constructor
         * @param {repository.ITemplate=} [properties] Properties to set
         */
        function Template(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Template id.
         * @member {string} id
         * @memberof repository.Template
         * @instance
         */
        Template.prototype.id = "";

        /**
         * Template mtime.
         * @member {number|Long} mtime
         * @memberof repository.Template
         * @instance
         */
        Template.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Template name.
         * @member {string} name
         * @memberof repository.Template
         * @instance
         */
        Template.prototype.name = "";

        /**
         * Template extension.
         * @member {string} extension
         * @memberof repository.Template
         * @instance
         */
        Template.prototype.extension = "";

        /**
         * Template chunks.
         * @member {string} chunks
         * @memberof repository.Template
         * @instance
         */
        Template.prototype.chunks = "";

        /**
         * Template xxhash_checksum.
         * @member {string} xxhash_checksum
         * @memberof repository.Template
         * @instance
         */
        Template.prototype.xxhash_checksum = "";

        /**
         * Template file_size.
         * @member {number|Long} file_size
         * @memberof repository.Template
         * @instance
         */
        Template.prototype.file_size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Template trashed.
         * @member {boolean} trashed
         * @memberof repository.Template
         * @instance
         */
        Template.prototype.trashed = false;

        /**
         * Template synced.
         * @member {boolean} synced
         * @memberof repository.Template
         * @instance
         */
        Template.prototype.synced = false;

        /**
         * Creates a new Template instance using the specified properties.
         * @function create
         * @memberof repository.Template
         * @static
         * @param {repository.ITemplate=} [properties] Properties to set
         * @returns {repository.Template} Template instance
         */
        Template.create = function create(properties) {
            return new Template(properties);
        };

        /**
         * Encodes the specified Template message. Does not implicitly {@link repository.Template.verify|verify} messages.
         * @function encode
         * @memberof repository.Template
         * @static
         * @param {repository.ITemplate} message Template message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Template.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.name);
            if (message.extension != null && Object.hasOwnProperty.call(message, "extension"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.extension);
            if (message.chunks != null && Object.hasOwnProperty.call(message, "chunks"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.chunks);
            if (message.xxhash_checksum != null && Object.hasOwnProperty.call(message, "xxhash_checksum"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.xxhash_checksum);
            if (message.file_size != null && Object.hasOwnProperty.call(message, "file_size"))
                writer.uint32(/* id 7, wireType 0 =*/56).int64(message.file_size);
            if (message.trashed != null && Object.hasOwnProperty.call(message, "trashed"))
                writer.uint32(/* id 8, wireType 0 =*/64).bool(message.trashed);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 9, wireType 0 =*/72).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified Template message, length delimited. Does not implicitly {@link repository.Template.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Template
         * @static
         * @param {repository.ITemplate} message Template message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Template.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Template message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Template
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Template} Template
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Template.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Template();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.name = reader.string();
                        break;
                    }
                case 4: {
                        message.extension = reader.string();
                        break;
                    }
                case 5: {
                        message.chunks = reader.string();
                        break;
                    }
                case 6: {
                        message.xxhash_checksum = reader.string();
                        break;
                    }
                case 7: {
                        message.file_size = reader.int64();
                        break;
                    }
                case 8: {
                        message.trashed = reader.bool();
                        break;
                    }
                case 9: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Template message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Template
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Template} Template
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Template.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Template message.
         * @function verify
         * @memberof repository.Template
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Template.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.extension != null && message.hasOwnProperty("extension"))
                if (!$util.isString(message.extension))
                    return "extension: string expected";
            if (message.chunks != null && message.hasOwnProperty("chunks"))
                if (!$util.isString(message.chunks))
                    return "chunks: string expected";
            if (message.xxhash_checksum != null && message.hasOwnProperty("xxhash_checksum"))
                if (!$util.isString(message.xxhash_checksum))
                    return "xxhash_checksum: string expected";
            if (message.file_size != null && message.hasOwnProperty("file_size"))
                if (!$util.isInteger(message.file_size) && !(message.file_size && $util.isInteger(message.file_size.low) && $util.isInteger(message.file_size.high)))
                    return "file_size: integer|Long expected";
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                if (typeof message.trashed !== "boolean")
                    return "trashed: boolean expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a Template message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Template
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Template} Template
         */
        Template.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Template)
                return object;
            let message = new $root.repository.Template();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.name != null)
                message.name = String(object.name);
            if (object.extension != null)
                message.extension = String(object.extension);
            if (object.chunks != null)
                message.chunks = String(object.chunks);
            if (object.xxhash_checksum != null)
                message.xxhash_checksum = String(object.xxhash_checksum);
            if (object.file_size != null)
                if ($util.Long)
                    (message.file_size = $util.Long.fromValue(object.file_size)).unsigned = false;
                else if (typeof object.file_size === "string")
                    message.file_size = parseInt(object.file_size, 10);
                else if (typeof object.file_size === "number")
                    message.file_size = object.file_size;
                else if (typeof object.file_size === "object")
                    message.file_size = new $util.LongBits(object.file_size.low >>> 0, object.file_size.high >>> 0).toNumber();
            if (object.trashed != null)
                message.trashed = Boolean(object.trashed);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a Template message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Template
         * @static
         * @param {repository.Template} message Template
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Template.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.name = "";
                object.extension = "";
                object.chunks = "";
                object.xxhash_checksum = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.file_size = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.file_size = options.longs === String ? "0" : 0;
                object.trashed = false;
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.extension != null && message.hasOwnProperty("extension"))
                object.extension = message.extension;
            if (message.chunks != null && message.hasOwnProperty("chunks"))
                object.chunks = message.chunks;
            if (message.xxhash_checksum != null && message.hasOwnProperty("xxhash_checksum"))
                object.xxhash_checksum = message.xxhash_checksum;
            if (message.file_size != null && message.hasOwnProperty("file_size"))
                if (typeof message.file_size === "number")
                    object.file_size = options.longs === String ? String(message.file_size) : message.file_size;
                else
                    object.file_size = options.longs === String ? $util.Long.prototype.toString.call(message.file_size) : options.longs === Number ? new $util.LongBits(message.file_size.low >>> 0, message.file_size.high >>> 0).toNumber() : message.file_size;
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                object.trashed = message.trashed;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this Template to JSON.
         * @function toJSON
         * @memberof repository.Template
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Template.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Template
         * @function getTypeUrl
         * @memberof repository.Template
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Template.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Template";
        };

        return Template;
    })();

    repository.Preview = (function() {

        /**
         * Properties of a Preview.
         * @memberof repository
         * @interface IPreview
         * @property {string|null} [hash] Preview hash
         * @property {Uint8Array|null} [preview] Preview preview
         * @property {string|null} [extension] Preview extension
         */

        /**
         * Constructs a new Preview.
         * @memberof repository
         * @classdesc Represents a Preview.
         * @implements IPreview
         * @constructor
         * @param {repository.IPreview=} [properties] Properties to set
         */
        function Preview(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Preview hash.
         * @member {string} hash
         * @memberof repository.Preview
         * @instance
         */
        Preview.prototype.hash = "";

        /**
         * Preview preview.
         * @member {Uint8Array} preview
         * @memberof repository.Preview
         * @instance
         */
        Preview.prototype.preview = $util.newBuffer([]);

        /**
         * Preview extension.
         * @member {string} extension
         * @memberof repository.Preview
         * @instance
         */
        Preview.prototype.extension = "";

        /**
         * Creates a new Preview instance using the specified properties.
         * @function create
         * @memberof repository.Preview
         * @static
         * @param {repository.IPreview=} [properties] Properties to set
         * @returns {repository.Preview} Preview instance
         */
        Preview.create = function create(properties) {
            return new Preview(properties);
        };

        /**
         * Encodes the specified Preview message. Does not implicitly {@link repository.Preview.verify|verify} messages.
         * @function encode
         * @memberof repository.Preview
         * @static
         * @param {repository.IPreview} message Preview message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Preview.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.hash != null && Object.hasOwnProperty.call(message, "hash"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.hash);
            if (message.preview != null && Object.hasOwnProperty.call(message, "preview"))
                writer.uint32(/* id 2, wireType 2 =*/18).bytes(message.preview);
            if (message.extension != null && Object.hasOwnProperty.call(message, "extension"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.extension);
            return writer;
        };

        /**
         * Encodes the specified Preview message, length delimited. Does not implicitly {@link repository.Preview.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Preview
         * @static
         * @param {repository.IPreview} message Preview message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Preview.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Preview message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Preview
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Preview} Preview
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Preview.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Preview();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.hash = reader.string();
                        break;
                    }
                case 2: {
                        message.preview = reader.bytes();
                        break;
                    }
                case 3: {
                        message.extension = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Preview message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Preview
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Preview} Preview
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Preview.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Preview message.
         * @function verify
         * @memberof repository.Preview
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Preview.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.hash != null && message.hasOwnProperty("hash"))
                if (!$util.isString(message.hash))
                    return "hash: string expected";
            if (message.preview != null && message.hasOwnProperty("preview"))
                if (!(message.preview && typeof message.preview.length === "number" || $util.isString(message.preview)))
                    return "preview: buffer expected";
            if (message.extension != null && message.hasOwnProperty("extension"))
                if (!$util.isString(message.extension))
                    return "extension: string expected";
            return null;
        };

        /**
         * Creates a Preview message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Preview
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Preview} Preview
         */
        Preview.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Preview)
                return object;
            let message = new $root.repository.Preview();
            if (object.hash != null)
                message.hash = String(object.hash);
            if (object.preview != null)
                if (typeof object.preview === "string")
                    $util.base64.decode(object.preview, message.preview = $util.newBuffer($util.base64.length(object.preview)), 0);
                else if (object.preview.length >= 0)
                    message.preview = object.preview;
            if (object.extension != null)
                message.extension = String(object.extension);
            return message;
        };

        /**
         * Creates a plain object from a Preview message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Preview
         * @static
         * @param {repository.Preview} message Preview
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Preview.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.hash = "";
                if (options.bytes === String)
                    object.preview = "";
                else {
                    object.preview = [];
                    if (options.bytes !== Array)
                        object.preview = $util.newBuffer(object.preview);
                }
                object.extension = "";
            }
            if (message.hash != null && message.hasOwnProperty("hash"))
                object.hash = message.hash;
            if (message.preview != null && message.hasOwnProperty("preview"))
                object.preview = options.bytes === String ? $util.base64.encode(message.preview, 0, message.preview.length) : options.bytes === Array ? Array.prototype.slice.call(message.preview) : message.preview;
            if (message.extension != null && message.hasOwnProperty("extension"))
                object.extension = message.extension;
            return object;
        };

        /**
         * Converts this Preview to JSON.
         * @function toJSON
         * @memberof repository.Preview
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Preview.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Preview
         * @function getTypeUrl
         * @memberof repository.Preview
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Preview.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Preview";
        };

        return Preview;
    })();

    repository.Tomb = (function() {

        /**
         * Properties of a Tomb.
         * @memberof repository
         * @interface ITomb
         * @property {string|null} [id] Tomb id
         * @property {number|Long|null} [mtime] Tomb mtime
         * @property {string|null} [table_name] Tomb table_name
         * @property {boolean|null} [synced] Tomb synced
         */

        /**
         * Constructs a new Tomb.
         * @memberof repository
         * @classdesc Represents a Tomb.
         * @implements ITomb
         * @constructor
         * @param {repository.ITomb=} [properties] Properties to set
         */
        function Tomb(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Tomb id.
         * @member {string} id
         * @memberof repository.Tomb
         * @instance
         */
        Tomb.prototype.id = "";

        /**
         * Tomb mtime.
         * @member {number|Long} mtime
         * @memberof repository.Tomb
         * @instance
         */
        Tomb.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Tomb table_name.
         * @member {string} table_name
         * @memberof repository.Tomb
         * @instance
         */
        Tomb.prototype.table_name = "";

        /**
         * Tomb synced.
         * @member {boolean} synced
         * @memberof repository.Tomb
         * @instance
         */
        Tomb.prototype.synced = false;

        /**
         * Creates a new Tomb instance using the specified properties.
         * @function create
         * @memberof repository.Tomb
         * @static
         * @param {repository.ITomb=} [properties] Properties to set
         * @returns {repository.Tomb} Tomb instance
         */
        Tomb.create = function create(properties) {
            return new Tomb(properties);
        };

        /**
         * Encodes the specified Tomb message. Does not implicitly {@link repository.Tomb.verify|verify} messages.
         * @function encode
         * @memberof repository.Tomb
         * @static
         * @param {repository.ITomb} message Tomb message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Tomb.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.table_name != null && Object.hasOwnProperty.call(message, "table_name"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.table_name);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 4, wireType 0 =*/32).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified Tomb message, length delimited. Does not implicitly {@link repository.Tomb.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Tomb
         * @static
         * @param {repository.ITomb} message Tomb message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Tomb.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Tomb message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Tomb
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Tomb} Tomb
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Tomb.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Tomb();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.table_name = reader.string();
                        break;
                    }
                case 4: {
                        message.synced = reader.bool();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Tomb message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Tomb
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Tomb} Tomb
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Tomb.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Tomb message.
         * @function verify
         * @memberof repository.Tomb
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Tomb.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.table_name != null && message.hasOwnProperty("table_name"))
                if (!$util.isString(message.table_name))
                    return "table_name: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a Tomb message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Tomb
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Tomb} Tomb
         */
        Tomb.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Tomb)
                return object;
            let message = new $root.repository.Tomb();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.table_name != null)
                message.table_name = String(object.table_name);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a Tomb message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Tomb
         * @static
         * @param {repository.Tomb} message Tomb
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Tomb.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.table_name = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.table_name != null && message.hasOwnProperty("table_name"))
                object.table_name = message.table_name;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this Tomb to JSON.
         * @function toJSON
         * @memberof repository.Tomb
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Tomb.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Tomb
         * @function getTypeUrl
         * @memberof repository.Tomb
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Tomb.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Tomb";
        };

        return Tomb;
    })();

    repository.ProjectData = (function() {

        /**
         * Properties of a ProjectData.
         * @memberof repository
         * @interface IProjectData
         * @property {string|null} [project_preview] ProjectData project_preview
         * @property {Array.<repository.ITask>|null} [tasks] ProjectData tasks
         * @property {Array.<repository.ITaskType>|null} [task_types] ProjectData task_types
         * @property {Array.<repository.ICheckpoint>|null} [tasks_checkpoints] ProjectData tasks_checkpoints
         * @property {Array.<repository.ITaskDependency>|null} [task_dependencies] ProjectData task_dependencies
         * @property {Array.<repository.IEntityDependency>|null} [entity_dependencies] ProjectData entity_dependencies
         * @property {Array.<repository.IStatus>|null} [statuses] ProjectData statuses
         * @property {Array.<repository.IDependencyType>|null} [dependency_types] ProjectData dependency_types
         * @property {Array.<repository.IUser>|null} [users] ProjectData users
         * @property {Array.<repository.IRole>|null} [roles] ProjectData roles
         * @property {Array.<repository.IEntityType>|null} [entity_types] ProjectData entity_types
         * @property {Array.<repository.IEntity>|null} [entities] ProjectData entities
         * @property {Array.<repository.IEntityAssignee>|null} [entity_assignees] ProjectData entity_assignees
         * @property {Array.<repository.ITemplate>|null} [templates] ProjectData templates
         * @property {Array.<repository.ITag>|null} [tags] ProjectData tags
         * @property {Array.<repository.ITaskTag>|null} [tasks_tags] ProjectData tasks_tags
         * @property {Array.<repository.IWorkflow>|null} [workflows] ProjectData workflows
         * @property {Array.<repository.IWorkflowLink>|null} [workflow_links] ProjectData workflow_links
         * @property {Array.<repository.IWorkflowEntity>|null} [workflow_entities] ProjectData workflow_entities
         * @property {Array.<repository.IWorkflowTask>|null} [workflow_tasks] ProjectData workflow_tasks
         * @property {Array.<repository.ITomb>|null} [tomb] ProjectData tomb
         */

        /**
         * Constructs a new ProjectData.
         * @memberof repository
         * @classdesc Represents a ProjectData.
         * @implements IProjectData
         * @constructor
         * @param {repository.IProjectData=} [properties] Properties to set
         */
        function ProjectData(properties) {
            this.tasks = [];
            this.task_types = [];
            this.tasks_checkpoints = [];
            this.task_dependencies = [];
            this.entity_dependencies = [];
            this.statuses = [];
            this.dependency_types = [];
            this.users = [];
            this.roles = [];
            this.entity_types = [];
            this.entities = [];
            this.entity_assignees = [];
            this.templates = [];
            this.tags = [];
            this.tasks_tags = [];
            this.workflows = [];
            this.workflow_links = [];
            this.workflow_entities = [];
            this.workflow_tasks = [];
            this.tomb = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ProjectData project_preview.
         * @member {string} project_preview
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.project_preview = "";

        /**
         * ProjectData tasks.
         * @member {Array.<repository.ITask>} tasks
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.tasks = $util.emptyArray;

        /**
         * ProjectData task_types.
         * @member {Array.<repository.ITaskType>} task_types
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.task_types = $util.emptyArray;

        /**
         * ProjectData tasks_checkpoints.
         * @member {Array.<repository.ICheckpoint>} tasks_checkpoints
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.tasks_checkpoints = $util.emptyArray;

        /**
         * ProjectData task_dependencies.
         * @member {Array.<repository.ITaskDependency>} task_dependencies
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.task_dependencies = $util.emptyArray;

        /**
         * ProjectData entity_dependencies.
         * @member {Array.<repository.IEntityDependency>} entity_dependencies
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.entity_dependencies = $util.emptyArray;

        /**
         * ProjectData statuses.
         * @member {Array.<repository.IStatus>} statuses
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.statuses = $util.emptyArray;

        /**
         * ProjectData dependency_types.
         * @member {Array.<repository.IDependencyType>} dependency_types
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.dependency_types = $util.emptyArray;

        /**
         * ProjectData users.
         * @member {Array.<repository.IUser>} users
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.users = $util.emptyArray;

        /**
         * ProjectData roles.
         * @member {Array.<repository.IRole>} roles
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.roles = $util.emptyArray;

        /**
         * ProjectData entity_types.
         * @member {Array.<repository.IEntityType>} entity_types
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.entity_types = $util.emptyArray;

        /**
         * ProjectData entities.
         * @member {Array.<repository.IEntity>} entities
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.entities = $util.emptyArray;

        /**
         * ProjectData entity_assignees.
         * @member {Array.<repository.IEntityAssignee>} entity_assignees
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.entity_assignees = $util.emptyArray;

        /**
         * ProjectData templates.
         * @member {Array.<repository.ITemplate>} templates
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.templates = $util.emptyArray;

        /**
         * ProjectData tags.
         * @member {Array.<repository.ITag>} tags
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.tags = $util.emptyArray;

        /**
         * ProjectData tasks_tags.
         * @member {Array.<repository.ITaskTag>} tasks_tags
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.tasks_tags = $util.emptyArray;

        /**
         * ProjectData workflows.
         * @member {Array.<repository.IWorkflow>} workflows
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.workflows = $util.emptyArray;

        /**
         * ProjectData workflow_links.
         * @member {Array.<repository.IWorkflowLink>} workflow_links
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.workflow_links = $util.emptyArray;

        /**
         * ProjectData workflow_entities.
         * @member {Array.<repository.IWorkflowEntity>} workflow_entities
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.workflow_entities = $util.emptyArray;

        /**
         * ProjectData workflow_tasks.
         * @member {Array.<repository.IWorkflowTask>} workflow_tasks
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.workflow_tasks = $util.emptyArray;

        /**
         * ProjectData tomb.
         * @member {Array.<repository.ITomb>} tomb
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.tomb = $util.emptyArray;

        /**
         * Creates a new ProjectData instance using the specified properties.
         * @function create
         * @memberof repository.ProjectData
         * @static
         * @param {repository.IProjectData=} [properties] Properties to set
         * @returns {repository.ProjectData} ProjectData instance
         */
        ProjectData.create = function create(properties) {
            return new ProjectData(properties);
        };

        /**
         * Encodes the specified ProjectData message. Does not implicitly {@link repository.ProjectData.verify|verify} messages.
         * @function encode
         * @memberof repository.ProjectData
         * @static
         * @param {repository.IProjectData} message ProjectData message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ProjectData.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.project_preview != null && Object.hasOwnProperty.call(message, "project_preview"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.project_preview);
            if (message.tasks != null && message.tasks.length)
                for (let i = 0; i < message.tasks.length; ++i)
                    $root.repository.Task.encode(message.tasks[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            if (message.task_types != null && message.task_types.length)
                for (let i = 0; i < message.task_types.length; ++i)
                    $root.repository.TaskType.encode(message.task_types[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            if (message.tasks_checkpoints != null && message.tasks_checkpoints.length)
                for (let i = 0; i < message.tasks_checkpoints.length; ++i)
                    $root.repository.Checkpoint.encode(message.tasks_checkpoints[i], writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
            if (message.task_dependencies != null && message.task_dependencies.length)
                for (let i = 0; i < message.task_dependencies.length; ++i)
                    $root.repository.TaskDependency.encode(message.task_dependencies[i], writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
            if (message.entity_dependencies != null && message.entity_dependencies.length)
                for (let i = 0; i < message.entity_dependencies.length; ++i)
                    $root.repository.EntityDependency.encode(message.entity_dependencies[i], writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
            if (message.statuses != null && message.statuses.length)
                for (let i = 0; i < message.statuses.length; ++i)
                    $root.repository.Status.encode(message.statuses[i], writer.uint32(/* id 7, wireType 2 =*/58).fork()).ldelim();
            if (message.dependency_types != null && message.dependency_types.length)
                for (let i = 0; i < message.dependency_types.length; ++i)
                    $root.repository.DependencyType.encode(message.dependency_types[i], writer.uint32(/* id 8, wireType 2 =*/66).fork()).ldelim();
            if (message.users != null && message.users.length)
                for (let i = 0; i < message.users.length; ++i)
                    $root.repository.User.encode(message.users[i], writer.uint32(/* id 9, wireType 2 =*/74).fork()).ldelim();
            if (message.roles != null && message.roles.length)
                for (let i = 0; i < message.roles.length; ++i)
                    $root.repository.Role.encode(message.roles[i], writer.uint32(/* id 10, wireType 2 =*/82).fork()).ldelim();
            if (message.entity_types != null && message.entity_types.length)
                for (let i = 0; i < message.entity_types.length; ++i)
                    $root.repository.EntityType.encode(message.entity_types[i], writer.uint32(/* id 11, wireType 2 =*/90).fork()).ldelim();
            if (message.entities != null && message.entities.length)
                for (let i = 0; i < message.entities.length; ++i)
                    $root.repository.Entity.encode(message.entities[i], writer.uint32(/* id 12, wireType 2 =*/98).fork()).ldelim();
            if (message.entity_assignees != null && message.entity_assignees.length)
                for (let i = 0; i < message.entity_assignees.length; ++i)
                    $root.repository.EntityAssignee.encode(message.entity_assignees[i], writer.uint32(/* id 13, wireType 2 =*/106).fork()).ldelim();
            if (message.templates != null && message.templates.length)
                for (let i = 0; i < message.templates.length; ++i)
                    $root.repository.Template.encode(message.templates[i], writer.uint32(/* id 14, wireType 2 =*/114).fork()).ldelim();
            if (message.tags != null && message.tags.length)
                for (let i = 0; i < message.tags.length; ++i)
                    $root.repository.Tag.encode(message.tags[i], writer.uint32(/* id 15, wireType 2 =*/122).fork()).ldelim();
            if (message.tasks_tags != null && message.tasks_tags.length)
                for (let i = 0; i < message.tasks_tags.length; ++i)
                    $root.repository.TaskTag.encode(message.tasks_tags[i], writer.uint32(/* id 16, wireType 2 =*/130).fork()).ldelim();
            if (message.workflows != null && message.workflows.length)
                for (let i = 0; i < message.workflows.length; ++i)
                    $root.repository.Workflow.encode(message.workflows[i], writer.uint32(/* id 17, wireType 2 =*/138).fork()).ldelim();
            if (message.workflow_links != null && message.workflow_links.length)
                for (let i = 0; i < message.workflow_links.length; ++i)
                    $root.repository.WorkflowLink.encode(message.workflow_links[i], writer.uint32(/* id 18, wireType 2 =*/146).fork()).ldelim();
            if (message.workflow_entities != null && message.workflow_entities.length)
                for (let i = 0; i < message.workflow_entities.length; ++i)
                    $root.repository.WorkflowEntity.encode(message.workflow_entities[i], writer.uint32(/* id 19, wireType 2 =*/154).fork()).ldelim();
            if (message.workflow_tasks != null && message.workflow_tasks.length)
                for (let i = 0; i < message.workflow_tasks.length; ++i)
                    $root.repository.WorkflowTask.encode(message.workflow_tasks[i], writer.uint32(/* id 20, wireType 2 =*/162).fork()).ldelim();
            if (message.tomb != null && message.tomb.length)
                for (let i = 0; i < message.tomb.length; ++i)
                    $root.repository.Tomb.encode(message.tomb[i], writer.uint32(/* id 21, wireType 2 =*/170).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified ProjectData message, length delimited. Does not implicitly {@link repository.ProjectData.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.ProjectData
         * @static
         * @param {repository.IProjectData} message ProjectData message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ProjectData.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a ProjectData message from the specified reader or buffer.
         * @function decode
         * @memberof repository.ProjectData
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.ProjectData} ProjectData
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ProjectData.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.ProjectData();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.project_preview = reader.string();
                        break;
                    }
                case 2: {
                        if (!(message.tasks && message.tasks.length))
                            message.tasks = [];
                        message.tasks.push($root.repository.Task.decode(reader, reader.uint32()));
                        break;
                    }
                case 3: {
                        if (!(message.task_types && message.task_types.length))
                            message.task_types = [];
                        message.task_types.push($root.repository.TaskType.decode(reader, reader.uint32()));
                        break;
                    }
                case 4: {
                        if (!(message.tasks_checkpoints && message.tasks_checkpoints.length))
                            message.tasks_checkpoints = [];
                        message.tasks_checkpoints.push($root.repository.Checkpoint.decode(reader, reader.uint32()));
                        break;
                    }
                case 5: {
                        if (!(message.task_dependencies && message.task_dependencies.length))
                            message.task_dependencies = [];
                        message.task_dependencies.push($root.repository.TaskDependency.decode(reader, reader.uint32()));
                        break;
                    }
                case 6: {
                        if (!(message.entity_dependencies && message.entity_dependencies.length))
                            message.entity_dependencies = [];
                        message.entity_dependencies.push($root.repository.EntityDependency.decode(reader, reader.uint32()));
                        break;
                    }
                case 7: {
                        if (!(message.statuses && message.statuses.length))
                            message.statuses = [];
                        message.statuses.push($root.repository.Status.decode(reader, reader.uint32()));
                        break;
                    }
                case 8: {
                        if (!(message.dependency_types && message.dependency_types.length))
                            message.dependency_types = [];
                        message.dependency_types.push($root.repository.DependencyType.decode(reader, reader.uint32()));
                        break;
                    }
                case 9: {
                        if (!(message.users && message.users.length))
                            message.users = [];
                        message.users.push($root.repository.User.decode(reader, reader.uint32()));
                        break;
                    }
                case 10: {
                        if (!(message.roles && message.roles.length))
                            message.roles = [];
                        message.roles.push($root.repository.Role.decode(reader, reader.uint32()));
                        break;
                    }
                case 11: {
                        if (!(message.entity_types && message.entity_types.length))
                            message.entity_types = [];
                        message.entity_types.push($root.repository.EntityType.decode(reader, reader.uint32()));
                        break;
                    }
                case 12: {
                        if (!(message.entities && message.entities.length))
                            message.entities = [];
                        message.entities.push($root.repository.Entity.decode(reader, reader.uint32()));
                        break;
                    }
                case 13: {
                        if (!(message.entity_assignees && message.entity_assignees.length))
                            message.entity_assignees = [];
                        message.entity_assignees.push($root.repository.EntityAssignee.decode(reader, reader.uint32()));
                        break;
                    }
                case 14: {
                        if (!(message.templates && message.templates.length))
                            message.templates = [];
                        message.templates.push($root.repository.Template.decode(reader, reader.uint32()));
                        break;
                    }
                case 15: {
                        if (!(message.tags && message.tags.length))
                            message.tags = [];
                        message.tags.push($root.repository.Tag.decode(reader, reader.uint32()));
                        break;
                    }
                case 16: {
                        if (!(message.tasks_tags && message.tasks_tags.length))
                            message.tasks_tags = [];
                        message.tasks_tags.push($root.repository.TaskTag.decode(reader, reader.uint32()));
                        break;
                    }
                case 17: {
                        if (!(message.workflows && message.workflows.length))
                            message.workflows = [];
                        message.workflows.push($root.repository.Workflow.decode(reader, reader.uint32()));
                        break;
                    }
                case 18: {
                        if (!(message.workflow_links && message.workflow_links.length))
                            message.workflow_links = [];
                        message.workflow_links.push($root.repository.WorkflowLink.decode(reader, reader.uint32()));
                        break;
                    }
                case 19: {
                        if (!(message.workflow_entities && message.workflow_entities.length))
                            message.workflow_entities = [];
                        message.workflow_entities.push($root.repository.WorkflowEntity.decode(reader, reader.uint32()));
                        break;
                    }
                case 20: {
                        if (!(message.workflow_tasks && message.workflow_tasks.length))
                            message.workflow_tasks = [];
                        message.workflow_tasks.push($root.repository.WorkflowTask.decode(reader, reader.uint32()));
                        break;
                    }
                case 21: {
                        if (!(message.tomb && message.tomb.length))
                            message.tomb = [];
                        message.tomb.push($root.repository.Tomb.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a ProjectData message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.ProjectData
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.ProjectData} ProjectData
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ProjectData.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a ProjectData message.
         * @function verify
         * @memberof repository.ProjectData
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        ProjectData.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.project_preview != null && message.hasOwnProperty("project_preview"))
                if (!$util.isString(message.project_preview))
                    return "project_preview: string expected";
            if (message.tasks != null && message.hasOwnProperty("tasks")) {
                if (!Array.isArray(message.tasks))
                    return "tasks: array expected";
                for (let i = 0; i < message.tasks.length; ++i) {
                    let error = $root.repository.Task.verify(message.tasks[i]);
                    if (error)
                        return "tasks." + error;
                }
            }
            if (message.task_types != null && message.hasOwnProperty("task_types")) {
                if (!Array.isArray(message.task_types))
                    return "task_types: array expected";
                for (let i = 0; i < message.task_types.length; ++i) {
                    let error = $root.repository.TaskType.verify(message.task_types[i]);
                    if (error)
                        return "task_types." + error;
                }
            }
            if (message.tasks_checkpoints != null && message.hasOwnProperty("tasks_checkpoints")) {
                if (!Array.isArray(message.tasks_checkpoints))
                    return "tasks_checkpoints: array expected";
                for (let i = 0; i < message.tasks_checkpoints.length; ++i) {
                    let error = $root.repository.Checkpoint.verify(message.tasks_checkpoints[i]);
                    if (error)
                        return "tasks_checkpoints." + error;
                }
            }
            if (message.task_dependencies != null && message.hasOwnProperty("task_dependencies")) {
                if (!Array.isArray(message.task_dependencies))
                    return "task_dependencies: array expected";
                for (let i = 0; i < message.task_dependencies.length; ++i) {
                    let error = $root.repository.TaskDependency.verify(message.task_dependencies[i]);
                    if (error)
                        return "task_dependencies." + error;
                }
            }
            if (message.entity_dependencies != null && message.hasOwnProperty("entity_dependencies")) {
                if (!Array.isArray(message.entity_dependencies))
                    return "entity_dependencies: array expected";
                for (let i = 0; i < message.entity_dependencies.length; ++i) {
                    let error = $root.repository.EntityDependency.verify(message.entity_dependencies[i]);
                    if (error)
                        return "entity_dependencies." + error;
                }
            }
            if (message.statuses != null && message.hasOwnProperty("statuses")) {
                if (!Array.isArray(message.statuses))
                    return "statuses: array expected";
                for (let i = 0; i < message.statuses.length; ++i) {
                    let error = $root.repository.Status.verify(message.statuses[i]);
                    if (error)
                        return "statuses." + error;
                }
            }
            if (message.dependency_types != null && message.hasOwnProperty("dependency_types")) {
                if (!Array.isArray(message.dependency_types))
                    return "dependency_types: array expected";
                for (let i = 0; i < message.dependency_types.length; ++i) {
                    let error = $root.repository.DependencyType.verify(message.dependency_types[i]);
                    if (error)
                        return "dependency_types." + error;
                }
            }
            if (message.users != null && message.hasOwnProperty("users")) {
                if (!Array.isArray(message.users))
                    return "users: array expected";
                for (let i = 0; i < message.users.length; ++i) {
                    let error = $root.repository.User.verify(message.users[i]);
                    if (error)
                        return "users." + error;
                }
            }
            if (message.roles != null && message.hasOwnProperty("roles")) {
                if (!Array.isArray(message.roles))
                    return "roles: array expected";
                for (let i = 0; i < message.roles.length; ++i) {
                    let error = $root.repository.Role.verify(message.roles[i]);
                    if (error)
                        return "roles." + error;
                }
            }
            if (message.entity_types != null && message.hasOwnProperty("entity_types")) {
                if (!Array.isArray(message.entity_types))
                    return "entity_types: array expected";
                for (let i = 0; i < message.entity_types.length; ++i) {
                    let error = $root.repository.EntityType.verify(message.entity_types[i]);
                    if (error)
                        return "entity_types." + error;
                }
            }
            if (message.entities != null && message.hasOwnProperty("entities")) {
                if (!Array.isArray(message.entities))
                    return "entities: array expected";
                for (let i = 0; i < message.entities.length; ++i) {
                    let error = $root.repository.Entity.verify(message.entities[i]);
                    if (error)
                        return "entities." + error;
                }
            }
            if (message.entity_assignees != null && message.hasOwnProperty("entity_assignees")) {
                if (!Array.isArray(message.entity_assignees))
                    return "entity_assignees: array expected";
                for (let i = 0; i < message.entity_assignees.length; ++i) {
                    let error = $root.repository.EntityAssignee.verify(message.entity_assignees[i]);
                    if (error)
                        return "entity_assignees." + error;
                }
            }
            if (message.templates != null && message.hasOwnProperty("templates")) {
                if (!Array.isArray(message.templates))
                    return "templates: array expected";
                for (let i = 0; i < message.templates.length; ++i) {
                    let error = $root.repository.Template.verify(message.templates[i]);
                    if (error)
                        return "templates." + error;
                }
            }
            if (message.tags != null && message.hasOwnProperty("tags")) {
                if (!Array.isArray(message.tags))
                    return "tags: array expected";
                for (let i = 0; i < message.tags.length; ++i) {
                    let error = $root.repository.Tag.verify(message.tags[i]);
                    if (error)
                        return "tags." + error;
                }
            }
            if (message.tasks_tags != null && message.hasOwnProperty("tasks_tags")) {
                if (!Array.isArray(message.tasks_tags))
                    return "tasks_tags: array expected";
                for (let i = 0; i < message.tasks_tags.length; ++i) {
                    let error = $root.repository.TaskTag.verify(message.tasks_tags[i]);
                    if (error)
                        return "tasks_tags." + error;
                }
            }
            if (message.workflows != null && message.hasOwnProperty("workflows")) {
                if (!Array.isArray(message.workflows))
                    return "workflows: array expected";
                for (let i = 0; i < message.workflows.length; ++i) {
                    let error = $root.repository.Workflow.verify(message.workflows[i]);
                    if (error)
                        return "workflows." + error;
                }
            }
            if (message.workflow_links != null && message.hasOwnProperty("workflow_links")) {
                if (!Array.isArray(message.workflow_links))
                    return "workflow_links: array expected";
                for (let i = 0; i < message.workflow_links.length; ++i) {
                    let error = $root.repository.WorkflowLink.verify(message.workflow_links[i]);
                    if (error)
                        return "workflow_links." + error;
                }
            }
            if (message.workflow_entities != null && message.hasOwnProperty("workflow_entities")) {
                if (!Array.isArray(message.workflow_entities))
                    return "workflow_entities: array expected";
                for (let i = 0; i < message.workflow_entities.length; ++i) {
                    let error = $root.repository.WorkflowEntity.verify(message.workflow_entities[i]);
                    if (error)
                        return "workflow_entities." + error;
                }
            }
            if (message.workflow_tasks != null && message.hasOwnProperty("workflow_tasks")) {
                if (!Array.isArray(message.workflow_tasks))
                    return "workflow_tasks: array expected";
                for (let i = 0; i < message.workflow_tasks.length; ++i) {
                    let error = $root.repository.WorkflowTask.verify(message.workflow_tasks[i]);
                    if (error)
                        return "workflow_tasks." + error;
                }
            }
            if (message.tomb != null && message.hasOwnProperty("tomb")) {
                if (!Array.isArray(message.tomb))
                    return "tomb: array expected";
                for (let i = 0; i < message.tomb.length; ++i) {
                    let error = $root.repository.Tomb.verify(message.tomb[i]);
                    if (error)
                        return "tomb." + error;
                }
            }
            return null;
        };

        /**
         * Creates a ProjectData message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.ProjectData
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.ProjectData} ProjectData
         */
        ProjectData.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.ProjectData)
                return object;
            let message = new $root.repository.ProjectData();
            if (object.project_preview != null)
                message.project_preview = String(object.project_preview);
            if (object.tasks) {
                if (!Array.isArray(object.tasks))
                    throw TypeError(".repository.ProjectData.tasks: array expected");
                message.tasks = [];
                for (let i = 0; i < object.tasks.length; ++i) {
                    if (typeof object.tasks[i] !== "object")
                        throw TypeError(".repository.ProjectData.tasks: object expected");
                    message.tasks[i] = $root.repository.Task.fromObject(object.tasks[i]);
                }
            }
            if (object.task_types) {
                if (!Array.isArray(object.task_types))
                    throw TypeError(".repository.ProjectData.task_types: array expected");
                message.task_types = [];
                for (let i = 0; i < object.task_types.length; ++i) {
                    if (typeof object.task_types[i] !== "object")
                        throw TypeError(".repository.ProjectData.task_types: object expected");
                    message.task_types[i] = $root.repository.TaskType.fromObject(object.task_types[i]);
                }
            }
            if (object.tasks_checkpoints) {
                if (!Array.isArray(object.tasks_checkpoints))
                    throw TypeError(".repository.ProjectData.tasks_checkpoints: array expected");
                message.tasks_checkpoints = [];
                for (let i = 0; i < object.tasks_checkpoints.length; ++i) {
                    if (typeof object.tasks_checkpoints[i] !== "object")
                        throw TypeError(".repository.ProjectData.tasks_checkpoints: object expected");
                    message.tasks_checkpoints[i] = $root.repository.Checkpoint.fromObject(object.tasks_checkpoints[i]);
                }
            }
            if (object.task_dependencies) {
                if (!Array.isArray(object.task_dependencies))
                    throw TypeError(".repository.ProjectData.task_dependencies: array expected");
                message.task_dependencies = [];
                for (let i = 0; i < object.task_dependencies.length; ++i) {
                    if (typeof object.task_dependencies[i] !== "object")
                        throw TypeError(".repository.ProjectData.task_dependencies: object expected");
                    message.task_dependencies[i] = $root.repository.TaskDependency.fromObject(object.task_dependencies[i]);
                }
            }
            if (object.entity_dependencies) {
                if (!Array.isArray(object.entity_dependencies))
                    throw TypeError(".repository.ProjectData.entity_dependencies: array expected");
                message.entity_dependencies = [];
                for (let i = 0; i < object.entity_dependencies.length; ++i) {
                    if (typeof object.entity_dependencies[i] !== "object")
                        throw TypeError(".repository.ProjectData.entity_dependencies: object expected");
                    message.entity_dependencies[i] = $root.repository.EntityDependency.fromObject(object.entity_dependencies[i]);
                }
            }
            if (object.statuses) {
                if (!Array.isArray(object.statuses))
                    throw TypeError(".repository.ProjectData.statuses: array expected");
                message.statuses = [];
                for (let i = 0; i < object.statuses.length; ++i) {
                    if (typeof object.statuses[i] !== "object")
                        throw TypeError(".repository.ProjectData.statuses: object expected");
                    message.statuses[i] = $root.repository.Status.fromObject(object.statuses[i]);
                }
            }
            if (object.dependency_types) {
                if (!Array.isArray(object.dependency_types))
                    throw TypeError(".repository.ProjectData.dependency_types: array expected");
                message.dependency_types = [];
                for (let i = 0; i < object.dependency_types.length; ++i) {
                    if (typeof object.dependency_types[i] !== "object")
                        throw TypeError(".repository.ProjectData.dependency_types: object expected");
                    message.dependency_types[i] = $root.repository.DependencyType.fromObject(object.dependency_types[i]);
                }
            }
            if (object.users) {
                if (!Array.isArray(object.users))
                    throw TypeError(".repository.ProjectData.users: array expected");
                message.users = [];
                for (let i = 0; i < object.users.length; ++i) {
                    if (typeof object.users[i] !== "object")
                        throw TypeError(".repository.ProjectData.users: object expected");
                    message.users[i] = $root.repository.User.fromObject(object.users[i]);
                }
            }
            if (object.roles) {
                if (!Array.isArray(object.roles))
                    throw TypeError(".repository.ProjectData.roles: array expected");
                message.roles = [];
                for (let i = 0; i < object.roles.length; ++i) {
                    if (typeof object.roles[i] !== "object")
                        throw TypeError(".repository.ProjectData.roles: object expected");
                    message.roles[i] = $root.repository.Role.fromObject(object.roles[i]);
                }
            }
            if (object.entity_types) {
                if (!Array.isArray(object.entity_types))
                    throw TypeError(".repository.ProjectData.entity_types: array expected");
                message.entity_types = [];
                for (let i = 0; i < object.entity_types.length; ++i) {
                    if (typeof object.entity_types[i] !== "object")
                        throw TypeError(".repository.ProjectData.entity_types: object expected");
                    message.entity_types[i] = $root.repository.EntityType.fromObject(object.entity_types[i]);
                }
            }
            if (object.entities) {
                if (!Array.isArray(object.entities))
                    throw TypeError(".repository.ProjectData.entities: array expected");
                message.entities = [];
                for (let i = 0; i < object.entities.length; ++i) {
                    if (typeof object.entities[i] !== "object")
                        throw TypeError(".repository.ProjectData.entities: object expected");
                    message.entities[i] = $root.repository.Entity.fromObject(object.entities[i]);
                }
            }
            if (object.entity_assignees) {
                if (!Array.isArray(object.entity_assignees))
                    throw TypeError(".repository.ProjectData.entity_assignees: array expected");
                message.entity_assignees = [];
                for (let i = 0; i < object.entity_assignees.length; ++i) {
                    if (typeof object.entity_assignees[i] !== "object")
                        throw TypeError(".repository.ProjectData.entity_assignees: object expected");
                    message.entity_assignees[i] = $root.repository.EntityAssignee.fromObject(object.entity_assignees[i]);
                }
            }
            if (object.templates) {
                if (!Array.isArray(object.templates))
                    throw TypeError(".repository.ProjectData.templates: array expected");
                message.templates = [];
                for (let i = 0; i < object.templates.length; ++i) {
                    if (typeof object.templates[i] !== "object")
                        throw TypeError(".repository.ProjectData.templates: object expected");
                    message.templates[i] = $root.repository.Template.fromObject(object.templates[i]);
                }
            }
            if (object.tags) {
                if (!Array.isArray(object.tags))
                    throw TypeError(".repository.ProjectData.tags: array expected");
                message.tags = [];
                for (let i = 0; i < object.tags.length; ++i) {
                    if (typeof object.tags[i] !== "object")
                        throw TypeError(".repository.ProjectData.tags: object expected");
                    message.tags[i] = $root.repository.Tag.fromObject(object.tags[i]);
                }
            }
            if (object.tasks_tags) {
                if (!Array.isArray(object.tasks_tags))
                    throw TypeError(".repository.ProjectData.tasks_tags: array expected");
                message.tasks_tags = [];
                for (let i = 0; i < object.tasks_tags.length; ++i) {
                    if (typeof object.tasks_tags[i] !== "object")
                        throw TypeError(".repository.ProjectData.tasks_tags: object expected");
                    message.tasks_tags[i] = $root.repository.TaskTag.fromObject(object.tasks_tags[i]);
                }
            }
            if (object.workflows) {
                if (!Array.isArray(object.workflows))
                    throw TypeError(".repository.ProjectData.workflows: array expected");
                message.workflows = [];
                for (let i = 0; i < object.workflows.length; ++i) {
                    if (typeof object.workflows[i] !== "object")
                        throw TypeError(".repository.ProjectData.workflows: object expected");
                    message.workflows[i] = $root.repository.Workflow.fromObject(object.workflows[i]);
                }
            }
            if (object.workflow_links) {
                if (!Array.isArray(object.workflow_links))
                    throw TypeError(".repository.ProjectData.workflow_links: array expected");
                message.workflow_links = [];
                for (let i = 0; i < object.workflow_links.length; ++i) {
                    if (typeof object.workflow_links[i] !== "object")
                        throw TypeError(".repository.ProjectData.workflow_links: object expected");
                    message.workflow_links[i] = $root.repository.WorkflowLink.fromObject(object.workflow_links[i]);
                }
            }
            if (object.workflow_entities) {
                if (!Array.isArray(object.workflow_entities))
                    throw TypeError(".repository.ProjectData.workflow_entities: array expected");
                message.workflow_entities = [];
                for (let i = 0; i < object.workflow_entities.length; ++i) {
                    if (typeof object.workflow_entities[i] !== "object")
                        throw TypeError(".repository.ProjectData.workflow_entities: object expected");
                    message.workflow_entities[i] = $root.repository.WorkflowEntity.fromObject(object.workflow_entities[i]);
                }
            }
            if (object.workflow_tasks) {
                if (!Array.isArray(object.workflow_tasks))
                    throw TypeError(".repository.ProjectData.workflow_tasks: array expected");
                message.workflow_tasks = [];
                for (let i = 0; i < object.workflow_tasks.length; ++i) {
                    if (typeof object.workflow_tasks[i] !== "object")
                        throw TypeError(".repository.ProjectData.workflow_tasks: object expected");
                    message.workflow_tasks[i] = $root.repository.WorkflowTask.fromObject(object.workflow_tasks[i]);
                }
            }
            if (object.tomb) {
                if (!Array.isArray(object.tomb))
                    throw TypeError(".repository.ProjectData.tomb: array expected");
                message.tomb = [];
                for (let i = 0; i < object.tomb.length; ++i) {
                    if (typeof object.tomb[i] !== "object")
                        throw TypeError(".repository.ProjectData.tomb: object expected");
                    message.tomb[i] = $root.repository.Tomb.fromObject(object.tomb[i]);
                }
            }
            return message;
        };

        /**
         * Creates a plain object from a ProjectData message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.ProjectData
         * @static
         * @param {repository.ProjectData} message ProjectData
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        ProjectData.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults) {
                object.tasks = [];
                object.task_types = [];
                object.tasks_checkpoints = [];
                object.task_dependencies = [];
                object.entity_dependencies = [];
                object.statuses = [];
                object.dependency_types = [];
                object.users = [];
                object.roles = [];
                object.entity_types = [];
                object.entities = [];
                object.entity_assignees = [];
                object.templates = [];
                object.tags = [];
                object.tasks_tags = [];
                object.workflows = [];
                object.workflow_links = [];
                object.workflow_entities = [];
                object.workflow_tasks = [];
                object.tomb = [];
            }
            if (options.defaults)
                object.project_preview = "";
            if (message.project_preview != null && message.hasOwnProperty("project_preview"))
                object.project_preview = message.project_preview;
            if (message.tasks && message.tasks.length) {
                object.tasks = [];
                for (let j = 0; j < message.tasks.length; ++j)
                    object.tasks[j] = $root.repository.Task.toObject(message.tasks[j], options);
            }
            if (message.task_types && message.task_types.length) {
                object.task_types = [];
                for (let j = 0; j < message.task_types.length; ++j)
                    object.task_types[j] = $root.repository.TaskType.toObject(message.task_types[j], options);
            }
            if (message.tasks_checkpoints && message.tasks_checkpoints.length) {
                object.tasks_checkpoints = [];
                for (let j = 0; j < message.tasks_checkpoints.length; ++j)
                    object.tasks_checkpoints[j] = $root.repository.Checkpoint.toObject(message.tasks_checkpoints[j], options);
            }
            if (message.task_dependencies && message.task_dependencies.length) {
                object.task_dependencies = [];
                for (let j = 0; j < message.task_dependencies.length; ++j)
                    object.task_dependencies[j] = $root.repository.TaskDependency.toObject(message.task_dependencies[j], options);
            }
            if (message.entity_dependencies && message.entity_dependencies.length) {
                object.entity_dependencies = [];
                for (let j = 0; j < message.entity_dependencies.length; ++j)
                    object.entity_dependencies[j] = $root.repository.EntityDependency.toObject(message.entity_dependencies[j], options);
            }
            if (message.statuses && message.statuses.length) {
                object.statuses = [];
                for (let j = 0; j < message.statuses.length; ++j)
                    object.statuses[j] = $root.repository.Status.toObject(message.statuses[j], options);
            }
            if (message.dependency_types && message.dependency_types.length) {
                object.dependency_types = [];
                for (let j = 0; j < message.dependency_types.length; ++j)
                    object.dependency_types[j] = $root.repository.DependencyType.toObject(message.dependency_types[j], options);
            }
            if (message.users && message.users.length) {
                object.users = [];
                for (let j = 0; j < message.users.length; ++j)
                    object.users[j] = $root.repository.User.toObject(message.users[j], options);
            }
            if (message.roles && message.roles.length) {
                object.roles = [];
                for (let j = 0; j < message.roles.length; ++j)
                    object.roles[j] = $root.repository.Role.toObject(message.roles[j], options);
            }
            if (message.entity_types && message.entity_types.length) {
                object.entity_types = [];
                for (let j = 0; j < message.entity_types.length; ++j)
                    object.entity_types[j] = $root.repository.EntityType.toObject(message.entity_types[j], options);
            }
            if (message.entities && message.entities.length) {
                object.entities = [];
                for (let j = 0; j < message.entities.length; ++j)
                    object.entities[j] = $root.repository.Entity.toObject(message.entities[j], options);
            }
            if (message.entity_assignees && message.entity_assignees.length) {
                object.entity_assignees = [];
                for (let j = 0; j < message.entity_assignees.length; ++j)
                    object.entity_assignees[j] = $root.repository.EntityAssignee.toObject(message.entity_assignees[j], options);
            }
            if (message.templates && message.templates.length) {
                object.templates = [];
                for (let j = 0; j < message.templates.length; ++j)
                    object.templates[j] = $root.repository.Template.toObject(message.templates[j], options);
            }
            if (message.tags && message.tags.length) {
                object.tags = [];
                for (let j = 0; j < message.tags.length; ++j)
                    object.tags[j] = $root.repository.Tag.toObject(message.tags[j], options);
            }
            if (message.tasks_tags && message.tasks_tags.length) {
                object.tasks_tags = [];
                for (let j = 0; j < message.tasks_tags.length; ++j)
                    object.tasks_tags[j] = $root.repository.TaskTag.toObject(message.tasks_tags[j], options);
            }
            if (message.workflows && message.workflows.length) {
                object.workflows = [];
                for (let j = 0; j < message.workflows.length; ++j)
                    object.workflows[j] = $root.repository.Workflow.toObject(message.workflows[j], options);
            }
            if (message.workflow_links && message.workflow_links.length) {
                object.workflow_links = [];
                for (let j = 0; j < message.workflow_links.length; ++j)
                    object.workflow_links[j] = $root.repository.WorkflowLink.toObject(message.workflow_links[j], options);
            }
            if (message.workflow_entities && message.workflow_entities.length) {
                object.workflow_entities = [];
                for (let j = 0; j < message.workflow_entities.length; ++j)
                    object.workflow_entities[j] = $root.repository.WorkflowEntity.toObject(message.workflow_entities[j], options);
            }
            if (message.workflow_tasks && message.workflow_tasks.length) {
                object.workflow_tasks = [];
                for (let j = 0; j < message.workflow_tasks.length; ++j)
                    object.workflow_tasks[j] = $root.repository.WorkflowTask.toObject(message.workflow_tasks[j], options);
            }
            if (message.tomb && message.tomb.length) {
                object.tomb = [];
                for (let j = 0; j < message.tomb.length; ++j)
                    object.tomb[j] = $root.repository.Tomb.toObject(message.tomb[j], options);
            }
            return object;
        };

        /**
         * Converts this ProjectData to JSON.
         * @function toJSON
         * @memberof repository.ProjectData
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        ProjectData.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for ProjectData
         * @function getTypeUrl
         * @memberof repository.ProjectData
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ProjectData.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.ProjectData";
        };

        return ProjectData;
    })();

    repository.FullTask = (function() {

        /**
         * Properties of a FullTask.
         * @memberof repository
         * @interface IFullTask
         * @property {string|null} [id] FullTask id
         * @property {number|Long|null} [mtime] FullTask mtime
         * @property {string|null} [created_at] FullTask created_at
         * @property {string|null} [name] FullTask name
         * @property {string|null} [description] FullTask description
         * @property {string|null} [extension] FullTask extension
         * @property {boolean|null} [is_resource] FullTask is_resource
         * @property {string|null} [status_id] FullTask status_id
         * @property {string|null} [status_short_name] FullTask status_short_name
         * @property {string|null} [task_type_id] FullTask task_type_id
         * @property {string|null} [task_type_name] FullTask task_type_name
         * @property {string|null} [task_type_icon] FullTask task_type_icon
         * @property {string|null} [entity_id] FullTask entity_id
         * @property {string|null} [entity_name] FullTask entity_name
         * @property {string|null} [entity_path] FullTask entity_path
         * @property {string|null} [task_path] FullTask task_path
         * @property {string|null} [assignee_id] FullTask assignee_id
         * @property {string|null} [assignee_email] FullTask assignee_email
         * @property {string|null} [assignee_name] FullTask assignee_name
         * @property {string|null} [assigner_id] FullTask assigner_id
         * @property {string|null} [assigner_email] FullTask assigner_email
         * @property {string|null} [assigner_name] FullTask assigner_name
         * @property {boolean|null} [is_dependency] FullTask is_dependency
         * @property {number|null} [dependency_level] FullTask dependency_level
         * @property {string|null} [file_path] FullTask file_path
         * @property {Array.<string>|null} [tags] FullTask tags
         * @property {string|null} [tags_raw] FullTask tags_raw
         * @property {Array.<string>|null} [entity_dependencies] FullTask entity_dependencies
         * @property {string|null} [entity_dependencies_raw] FullTask entity_dependencies_raw
         * @property {Array.<string>|null} [dependencies] FullTask dependencies
         * @property {string|null} [dependencies_raw] FullTask dependencies_raw
         * @property {string|null} [file_status] FullTask file_status
         * @property {repository.IStatus|null} [status] FullTask status
         * @property {boolean|null} [is_link] FullTask is_link
         * @property {string|null} [pointer] FullTask pointer
         * @property {string|null} [preview_id] FullTask preview_id
         * @property {Uint8Array|null} [preview] FullTask preview
         * @property {string|null} [preview_extension] FullTask preview_extension
         * @property {Array.<repository.ICheckpoint>|null} [checkpoints] FullTask checkpoints
         * @property {boolean|null} [trashed] FullTask trashed
         * @property {boolean|null} [synced] FullTask synced
         * @property {string|null} [type] FullTask type
         */

        /**
         * Constructs a new FullTask.
         * @memberof repository
         * @classdesc Represents a FullTask.
         * @implements IFullTask
         * @constructor
         * @param {repository.IFullTask=} [properties] Properties to set
         */
        function FullTask(properties) {
            this.tags = [];
            this.entity_dependencies = [];
            this.dependencies = [];
            this.checkpoints = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * FullTask id.
         * @member {string} id
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.id = "";

        /**
         * FullTask mtime.
         * @member {number|Long} mtime
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * FullTask created_at.
         * @member {string} created_at
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.created_at = "";

        /**
         * FullTask name.
         * @member {string} name
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.name = "";

        /**
         * FullTask description.
         * @member {string} description
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.description = "";

        /**
         * FullTask extension.
         * @member {string} extension
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.extension = "";

        /**
         * FullTask is_resource.
         * @member {boolean} is_resource
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.is_resource = false;

        /**
         * FullTask status_id.
         * @member {string} status_id
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.status_id = "";

        /**
         * FullTask status_short_name.
         * @member {string} status_short_name
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.status_short_name = "";

        /**
         * FullTask task_type_id.
         * @member {string} task_type_id
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.task_type_id = "";

        /**
         * FullTask task_type_name.
         * @member {string} task_type_name
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.task_type_name = "";

        /**
         * FullTask task_type_icon.
         * @member {string} task_type_icon
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.task_type_icon = "";

        /**
         * FullTask entity_id.
         * @member {string} entity_id
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.entity_id = "";

        /**
         * FullTask entity_name.
         * @member {string} entity_name
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.entity_name = "";

        /**
         * FullTask entity_path.
         * @member {string} entity_path
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.entity_path = "";

        /**
         * FullTask task_path.
         * @member {string} task_path
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.task_path = "";

        /**
         * FullTask assignee_id.
         * @member {string} assignee_id
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.assignee_id = "";

        /**
         * FullTask assignee_email.
         * @member {string} assignee_email
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.assignee_email = "";

        /**
         * FullTask assignee_name.
         * @member {string} assignee_name
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.assignee_name = "";

        /**
         * FullTask assigner_id.
         * @member {string} assigner_id
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.assigner_id = "";

        /**
         * FullTask assigner_email.
         * @member {string} assigner_email
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.assigner_email = "";

        /**
         * FullTask assigner_name.
         * @member {string} assigner_name
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.assigner_name = "";

        /**
         * FullTask is_dependency.
         * @member {boolean} is_dependency
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.is_dependency = false;

        /**
         * FullTask dependency_level.
         * @member {number} dependency_level
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.dependency_level = 0;

        /**
         * FullTask file_path.
         * @member {string} file_path
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.file_path = "";

        /**
         * FullTask tags.
         * @member {Array.<string>} tags
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.tags = $util.emptyArray;

        /**
         * FullTask tags_raw.
         * @member {string} tags_raw
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.tags_raw = "";

        /**
         * FullTask entity_dependencies.
         * @member {Array.<string>} entity_dependencies
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.entity_dependencies = $util.emptyArray;

        /**
         * FullTask entity_dependencies_raw.
         * @member {string} entity_dependencies_raw
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.entity_dependencies_raw = "";

        /**
         * FullTask dependencies.
         * @member {Array.<string>} dependencies
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.dependencies = $util.emptyArray;

        /**
         * FullTask dependencies_raw.
         * @member {string} dependencies_raw
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.dependencies_raw = "";

        /**
         * FullTask file_status.
         * @member {string} file_status
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.file_status = "";

        /**
         * FullTask status.
         * @member {repository.IStatus|null|undefined} status
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.status = null;

        /**
         * FullTask is_link.
         * @member {boolean} is_link
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.is_link = false;

        /**
         * FullTask pointer.
         * @member {string} pointer
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.pointer = "";

        /**
         * FullTask preview_id.
         * @member {string} preview_id
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.preview_id = "";

        /**
         * FullTask preview.
         * @member {Uint8Array} preview
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.preview = $util.newBuffer([]);

        /**
         * FullTask preview_extension.
         * @member {string} preview_extension
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.preview_extension = "";

        /**
         * FullTask checkpoints.
         * @member {Array.<repository.ICheckpoint>} checkpoints
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.checkpoints = $util.emptyArray;

        /**
         * FullTask trashed.
         * @member {boolean} trashed
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.trashed = false;

        /**
         * FullTask synced.
         * @member {boolean} synced
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.synced = false;

        /**
         * FullTask type.
         * @member {string} type
         * @memberof repository.FullTask
         * @instance
         */
        FullTask.prototype.type = "";

        /**
         * Creates a new FullTask instance using the specified properties.
         * @function create
         * @memberof repository.FullTask
         * @static
         * @param {repository.IFullTask=} [properties] Properties to set
         * @returns {repository.FullTask} FullTask instance
         */
        FullTask.create = function create(properties) {
            return new FullTask(properties);
        };

        /**
         * Encodes the specified FullTask message. Does not implicitly {@link repository.FullTask.verify|verify} messages.
         * @function encode
         * @memberof repository.FullTask
         * @static
         * @param {repository.IFullTask} message FullTask message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FullTask.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.created_at != null && Object.hasOwnProperty.call(message, "created_at"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.created_at);
            if (message.name != null && Object.hasOwnProperty.call(message, "name"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.name);
            if (message.description != null && Object.hasOwnProperty.call(message, "description"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.description);
            if (message.extension != null && Object.hasOwnProperty.call(message, "extension"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.extension);
            if (message.is_resource != null && Object.hasOwnProperty.call(message, "is_resource"))
                writer.uint32(/* id 7, wireType 0 =*/56).bool(message.is_resource);
            if (message.status_id != null && Object.hasOwnProperty.call(message, "status_id"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.status_id);
            if (message.status_short_name != null && Object.hasOwnProperty.call(message, "status_short_name"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.status_short_name);
            if (message.task_type_id != null && Object.hasOwnProperty.call(message, "task_type_id"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.task_type_id);
            if (message.task_type_name != null && Object.hasOwnProperty.call(message, "task_type_name"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.task_type_name);
            if (message.task_type_icon != null && Object.hasOwnProperty.call(message, "task_type_icon"))
                writer.uint32(/* id 12, wireType 2 =*/98).string(message.task_type_icon);
            if (message.entity_id != null && Object.hasOwnProperty.call(message, "entity_id"))
                writer.uint32(/* id 13, wireType 2 =*/106).string(message.entity_id);
            if (message.entity_name != null && Object.hasOwnProperty.call(message, "entity_name"))
                writer.uint32(/* id 14, wireType 2 =*/114).string(message.entity_name);
            if (message.entity_path != null && Object.hasOwnProperty.call(message, "entity_path"))
                writer.uint32(/* id 15, wireType 2 =*/122).string(message.entity_path);
            if (message.task_path != null && Object.hasOwnProperty.call(message, "task_path"))
                writer.uint32(/* id 16, wireType 2 =*/130).string(message.task_path);
            if (message.assignee_id != null && Object.hasOwnProperty.call(message, "assignee_id"))
                writer.uint32(/* id 17, wireType 2 =*/138).string(message.assignee_id);
            if (message.assignee_email != null && Object.hasOwnProperty.call(message, "assignee_email"))
                writer.uint32(/* id 18, wireType 2 =*/146).string(message.assignee_email);
            if (message.assignee_name != null && Object.hasOwnProperty.call(message, "assignee_name"))
                writer.uint32(/* id 19, wireType 2 =*/154).string(message.assignee_name);
            if (message.assigner_id != null && Object.hasOwnProperty.call(message, "assigner_id"))
                writer.uint32(/* id 20, wireType 2 =*/162).string(message.assigner_id);
            if (message.assigner_email != null && Object.hasOwnProperty.call(message, "assigner_email"))
                writer.uint32(/* id 21, wireType 2 =*/170).string(message.assigner_email);
            if (message.assigner_name != null && Object.hasOwnProperty.call(message, "assigner_name"))
                writer.uint32(/* id 22, wireType 2 =*/178).string(message.assigner_name);
            if (message.is_dependency != null && Object.hasOwnProperty.call(message, "is_dependency"))
                writer.uint32(/* id 23, wireType 0 =*/184).bool(message.is_dependency);
            if (message.dependency_level != null && Object.hasOwnProperty.call(message, "dependency_level"))
                writer.uint32(/* id 24, wireType 0 =*/192).int32(message.dependency_level);
            if (message.file_path != null && Object.hasOwnProperty.call(message, "file_path"))
                writer.uint32(/* id 25, wireType 2 =*/202).string(message.file_path);
            if (message.tags != null && message.tags.length)
                for (let i = 0; i < message.tags.length; ++i)
                    writer.uint32(/* id 26, wireType 2 =*/210).string(message.tags[i]);
            if (message.tags_raw != null && Object.hasOwnProperty.call(message, "tags_raw"))
                writer.uint32(/* id 27, wireType 2 =*/218).string(message.tags_raw);
            if (message.entity_dependencies != null && message.entity_dependencies.length)
                for (let i = 0; i < message.entity_dependencies.length; ++i)
                    writer.uint32(/* id 28, wireType 2 =*/226).string(message.entity_dependencies[i]);
            if (message.entity_dependencies_raw != null && Object.hasOwnProperty.call(message, "entity_dependencies_raw"))
                writer.uint32(/* id 29, wireType 2 =*/234).string(message.entity_dependencies_raw);
            if (message.dependencies != null && message.dependencies.length)
                for (let i = 0; i < message.dependencies.length; ++i)
                    writer.uint32(/* id 30, wireType 2 =*/242).string(message.dependencies[i]);
            if (message.dependencies_raw != null && Object.hasOwnProperty.call(message, "dependencies_raw"))
                writer.uint32(/* id 31, wireType 2 =*/250).string(message.dependencies_raw);
            if (message.file_status != null && Object.hasOwnProperty.call(message, "file_status"))
                writer.uint32(/* id 32, wireType 2 =*/258).string(message.file_status);
            if (message.status != null && Object.hasOwnProperty.call(message, "status"))
                $root.repository.Status.encode(message.status, writer.uint32(/* id 33, wireType 2 =*/266).fork()).ldelim();
            if (message.is_link != null && Object.hasOwnProperty.call(message, "is_link"))
                writer.uint32(/* id 34, wireType 0 =*/272).bool(message.is_link);
            if (message.pointer != null && Object.hasOwnProperty.call(message, "pointer"))
                writer.uint32(/* id 35, wireType 2 =*/282).string(message.pointer);
            if (message.preview_id != null && Object.hasOwnProperty.call(message, "preview_id"))
                writer.uint32(/* id 36, wireType 2 =*/290).string(message.preview_id);
            if (message.preview != null && Object.hasOwnProperty.call(message, "preview"))
                writer.uint32(/* id 37, wireType 2 =*/298).bytes(message.preview);
            if (message.preview_extension != null && Object.hasOwnProperty.call(message, "preview_extension"))
                writer.uint32(/* id 38, wireType 2 =*/306).string(message.preview_extension);
            if (message.checkpoints != null && message.checkpoints.length)
                for (let i = 0; i < message.checkpoints.length; ++i)
                    $root.repository.Checkpoint.encode(message.checkpoints[i], writer.uint32(/* id 39, wireType 2 =*/314).fork()).ldelim();
            if (message.trashed != null && Object.hasOwnProperty.call(message, "trashed"))
                writer.uint32(/* id 40, wireType 0 =*/320).bool(message.trashed);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 41, wireType 0 =*/328).bool(message.synced);
            if (message.type != null && Object.hasOwnProperty.call(message, "type"))
                writer.uint32(/* id 42, wireType 2 =*/338).string(message.type);
            return writer;
        };

        /**
         * Encodes the specified FullTask message, length delimited. Does not implicitly {@link repository.FullTask.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.FullTask
         * @static
         * @param {repository.IFullTask} message FullTask message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FullTask.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a FullTask message from the specified reader or buffer.
         * @function decode
         * @memberof repository.FullTask
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.FullTask} FullTask
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FullTask.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.FullTask();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        message.id = reader.string();
                        break;
                    }
                case 2: {
                        message.mtime = reader.int64();
                        break;
                    }
                case 3: {
                        message.created_at = reader.string();
                        break;
                    }
                case 4: {
                        message.name = reader.string();
                        break;
                    }
                case 5: {
                        message.description = reader.string();
                        break;
                    }
                case 6: {
                        message.extension = reader.string();
                        break;
                    }
                case 7: {
                        message.is_resource = reader.bool();
                        break;
                    }
                case 8: {
                        message.status_id = reader.string();
                        break;
                    }
                case 9: {
                        message.status_short_name = reader.string();
                        break;
                    }
                case 10: {
                        message.task_type_id = reader.string();
                        break;
                    }
                case 11: {
                        message.task_type_name = reader.string();
                        break;
                    }
                case 12: {
                        message.task_type_icon = reader.string();
                        break;
                    }
                case 13: {
                        message.entity_id = reader.string();
                        break;
                    }
                case 14: {
                        message.entity_name = reader.string();
                        break;
                    }
                case 15: {
                        message.entity_path = reader.string();
                        break;
                    }
                case 16: {
                        message.task_path = reader.string();
                        break;
                    }
                case 17: {
                        message.assignee_id = reader.string();
                        break;
                    }
                case 18: {
                        message.assignee_email = reader.string();
                        break;
                    }
                case 19: {
                        message.assignee_name = reader.string();
                        break;
                    }
                case 20: {
                        message.assigner_id = reader.string();
                        break;
                    }
                case 21: {
                        message.assigner_email = reader.string();
                        break;
                    }
                case 22: {
                        message.assigner_name = reader.string();
                        break;
                    }
                case 23: {
                        message.is_dependency = reader.bool();
                        break;
                    }
                case 24: {
                        message.dependency_level = reader.int32();
                        break;
                    }
                case 25: {
                        message.file_path = reader.string();
                        break;
                    }
                case 26: {
                        if (!(message.tags && message.tags.length))
                            message.tags = [];
                        message.tags.push(reader.string());
                        break;
                    }
                case 27: {
                        message.tags_raw = reader.string();
                        break;
                    }
                case 28: {
                        if (!(message.entity_dependencies && message.entity_dependencies.length))
                            message.entity_dependencies = [];
                        message.entity_dependencies.push(reader.string());
                        break;
                    }
                case 29: {
                        message.entity_dependencies_raw = reader.string();
                        break;
                    }
                case 30: {
                        if (!(message.dependencies && message.dependencies.length))
                            message.dependencies = [];
                        message.dependencies.push(reader.string());
                        break;
                    }
                case 31: {
                        message.dependencies_raw = reader.string();
                        break;
                    }
                case 32: {
                        message.file_status = reader.string();
                        break;
                    }
                case 33: {
                        message.status = $root.repository.Status.decode(reader, reader.uint32());
                        break;
                    }
                case 34: {
                        message.is_link = reader.bool();
                        break;
                    }
                case 35: {
                        message.pointer = reader.string();
                        break;
                    }
                case 36: {
                        message.preview_id = reader.string();
                        break;
                    }
                case 37: {
                        message.preview = reader.bytes();
                        break;
                    }
                case 38: {
                        message.preview_extension = reader.string();
                        break;
                    }
                case 39: {
                        if (!(message.checkpoints && message.checkpoints.length))
                            message.checkpoints = [];
                        message.checkpoints.push($root.repository.Checkpoint.decode(reader, reader.uint32()));
                        break;
                    }
                case 40: {
                        message.trashed = reader.bool();
                        break;
                    }
                case 41: {
                        message.synced = reader.bool();
                        break;
                    }
                case 42: {
                        message.type = reader.string();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a FullTask message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.FullTask
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.FullTask} FullTask
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FullTask.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a FullTask message.
         * @function verify
         * @memberof repository.FullTask
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        FullTask.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.created_at != null && message.hasOwnProperty("created_at"))
                if (!$util.isString(message.created_at))
                    return "created_at: string expected";
            if (message.name != null && message.hasOwnProperty("name"))
                if (!$util.isString(message.name))
                    return "name: string expected";
            if (message.description != null && message.hasOwnProperty("description"))
                if (!$util.isString(message.description))
                    return "description: string expected";
            if (message.extension != null && message.hasOwnProperty("extension"))
                if (!$util.isString(message.extension))
                    return "extension: string expected";
            if (message.is_resource != null && message.hasOwnProperty("is_resource"))
                if (typeof message.is_resource !== "boolean")
                    return "is_resource: boolean expected";
            if (message.status_id != null && message.hasOwnProperty("status_id"))
                if (!$util.isString(message.status_id))
                    return "status_id: string expected";
            if (message.status_short_name != null && message.hasOwnProperty("status_short_name"))
                if (!$util.isString(message.status_short_name))
                    return "status_short_name: string expected";
            if (message.task_type_id != null && message.hasOwnProperty("task_type_id"))
                if (!$util.isString(message.task_type_id))
                    return "task_type_id: string expected";
            if (message.task_type_name != null && message.hasOwnProperty("task_type_name"))
                if (!$util.isString(message.task_type_name))
                    return "task_type_name: string expected";
            if (message.task_type_icon != null && message.hasOwnProperty("task_type_icon"))
                if (!$util.isString(message.task_type_icon))
                    return "task_type_icon: string expected";
            if (message.entity_id != null && message.hasOwnProperty("entity_id"))
                if (!$util.isString(message.entity_id))
                    return "entity_id: string expected";
            if (message.entity_name != null && message.hasOwnProperty("entity_name"))
                if (!$util.isString(message.entity_name))
                    return "entity_name: string expected";
            if (message.entity_path != null && message.hasOwnProperty("entity_path"))
                if (!$util.isString(message.entity_path))
                    return "entity_path: string expected";
            if (message.task_path != null && message.hasOwnProperty("task_path"))
                if (!$util.isString(message.task_path))
                    return "task_path: string expected";
            if (message.assignee_id != null && message.hasOwnProperty("assignee_id"))
                if (!$util.isString(message.assignee_id))
                    return "assignee_id: string expected";
            if (message.assignee_email != null && message.hasOwnProperty("assignee_email"))
                if (!$util.isString(message.assignee_email))
                    return "assignee_email: string expected";
            if (message.assignee_name != null && message.hasOwnProperty("assignee_name"))
                if (!$util.isString(message.assignee_name))
                    return "assignee_name: string expected";
            if (message.assigner_id != null && message.hasOwnProperty("assigner_id"))
                if (!$util.isString(message.assigner_id))
                    return "assigner_id: string expected";
            if (message.assigner_email != null && message.hasOwnProperty("assigner_email"))
                if (!$util.isString(message.assigner_email))
                    return "assigner_email: string expected";
            if (message.assigner_name != null && message.hasOwnProperty("assigner_name"))
                if (!$util.isString(message.assigner_name))
                    return "assigner_name: string expected";
            if (message.is_dependency != null && message.hasOwnProperty("is_dependency"))
                if (typeof message.is_dependency !== "boolean")
                    return "is_dependency: boolean expected";
            if (message.dependency_level != null && message.hasOwnProperty("dependency_level"))
                if (!$util.isInteger(message.dependency_level))
                    return "dependency_level: integer expected";
            if (message.file_path != null && message.hasOwnProperty("file_path"))
                if (!$util.isString(message.file_path))
                    return "file_path: string expected";
            if (message.tags != null && message.hasOwnProperty("tags")) {
                if (!Array.isArray(message.tags))
                    return "tags: array expected";
                for (let i = 0; i < message.tags.length; ++i)
                    if (!$util.isString(message.tags[i]))
                        return "tags: string[] expected";
            }
            if (message.tags_raw != null && message.hasOwnProperty("tags_raw"))
                if (!$util.isString(message.tags_raw))
                    return "tags_raw: string expected";
            if (message.entity_dependencies != null && message.hasOwnProperty("entity_dependencies")) {
                if (!Array.isArray(message.entity_dependencies))
                    return "entity_dependencies: array expected";
                for (let i = 0; i < message.entity_dependencies.length; ++i)
                    if (!$util.isString(message.entity_dependencies[i]))
                        return "entity_dependencies: string[] expected";
            }
            if (message.entity_dependencies_raw != null && message.hasOwnProperty("entity_dependencies_raw"))
                if (!$util.isString(message.entity_dependencies_raw))
                    return "entity_dependencies_raw: string expected";
            if (message.dependencies != null && message.hasOwnProperty("dependencies")) {
                if (!Array.isArray(message.dependencies))
                    return "dependencies: array expected";
                for (let i = 0; i < message.dependencies.length; ++i)
                    if (!$util.isString(message.dependencies[i]))
                        return "dependencies: string[] expected";
            }
            if (message.dependencies_raw != null && message.hasOwnProperty("dependencies_raw"))
                if (!$util.isString(message.dependencies_raw))
                    return "dependencies_raw: string expected";
            if (message.file_status != null && message.hasOwnProperty("file_status"))
                if (!$util.isString(message.file_status))
                    return "file_status: string expected";
            if (message.status != null && message.hasOwnProperty("status")) {
                let error = $root.repository.Status.verify(message.status);
                if (error)
                    return "status." + error;
            }
            if (message.is_link != null && message.hasOwnProperty("is_link"))
                if (typeof message.is_link !== "boolean")
                    return "is_link: boolean expected";
            if (message.pointer != null && message.hasOwnProperty("pointer"))
                if (!$util.isString(message.pointer))
                    return "pointer: string expected";
            if (message.preview_id != null && message.hasOwnProperty("preview_id"))
                if (!$util.isString(message.preview_id))
                    return "preview_id: string expected";
            if (message.preview != null && message.hasOwnProperty("preview"))
                if (!(message.preview && typeof message.preview.length === "number" || $util.isString(message.preview)))
                    return "preview: buffer expected";
            if (message.preview_extension != null && message.hasOwnProperty("preview_extension"))
                if (!$util.isString(message.preview_extension))
                    return "preview_extension: string expected";
            if (message.checkpoints != null && message.hasOwnProperty("checkpoints")) {
                if (!Array.isArray(message.checkpoints))
                    return "checkpoints: array expected";
                for (let i = 0; i < message.checkpoints.length; ++i) {
                    let error = $root.repository.Checkpoint.verify(message.checkpoints[i]);
                    if (error)
                        return "checkpoints." + error;
                }
            }
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                if (typeof message.trashed !== "boolean")
                    return "trashed: boolean expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            if (message.type != null && message.hasOwnProperty("type"))
                if (!$util.isString(message.type))
                    return "type: string expected";
            return null;
        };

        /**
         * Creates a FullTask message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.FullTask
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.FullTask} FullTask
         */
        FullTask.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.FullTask)
                return object;
            let message = new $root.repository.FullTask();
            if (object.id != null)
                message.id = String(object.id);
            if (object.mtime != null)
                if ($util.Long)
                    (message.mtime = $util.Long.fromValue(object.mtime)).unsigned = false;
                else if (typeof object.mtime === "string")
                    message.mtime = parseInt(object.mtime, 10);
                else if (typeof object.mtime === "number")
                    message.mtime = object.mtime;
                else if (typeof object.mtime === "object")
                    message.mtime = new $util.LongBits(object.mtime.low >>> 0, object.mtime.high >>> 0).toNumber();
            if (object.created_at != null)
                message.created_at = String(object.created_at);
            if (object.name != null)
                message.name = String(object.name);
            if (object.description != null)
                message.description = String(object.description);
            if (object.extension != null)
                message.extension = String(object.extension);
            if (object.is_resource != null)
                message.is_resource = Boolean(object.is_resource);
            if (object.status_id != null)
                message.status_id = String(object.status_id);
            if (object.status_short_name != null)
                message.status_short_name = String(object.status_short_name);
            if (object.task_type_id != null)
                message.task_type_id = String(object.task_type_id);
            if (object.task_type_name != null)
                message.task_type_name = String(object.task_type_name);
            if (object.task_type_icon != null)
                message.task_type_icon = String(object.task_type_icon);
            if (object.entity_id != null)
                message.entity_id = String(object.entity_id);
            if (object.entity_name != null)
                message.entity_name = String(object.entity_name);
            if (object.entity_path != null)
                message.entity_path = String(object.entity_path);
            if (object.task_path != null)
                message.task_path = String(object.task_path);
            if (object.assignee_id != null)
                message.assignee_id = String(object.assignee_id);
            if (object.assignee_email != null)
                message.assignee_email = String(object.assignee_email);
            if (object.assignee_name != null)
                message.assignee_name = String(object.assignee_name);
            if (object.assigner_id != null)
                message.assigner_id = String(object.assigner_id);
            if (object.assigner_email != null)
                message.assigner_email = String(object.assigner_email);
            if (object.assigner_name != null)
                message.assigner_name = String(object.assigner_name);
            if (object.is_dependency != null)
                message.is_dependency = Boolean(object.is_dependency);
            if (object.dependency_level != null)
                message.dependency_level = object.dependency_level | 0;
            if (object.file_path != null)
                message.file_path = String(object.file_path);
            if (object.tags) {
                if (!Array.isArray(object.tags))
                    throw TypeError(".repository.FullTask.tags: array expected");
                message.tags = [];
                for (let i = 0; i < object.tags.length; ++i)
                    message.tags[i] = String(object.tags[i]);
            }
            if (object.tags_raw != null)
                message.tags_raw = String(object.tags_raw);
            if (object.entity_dependencies) {
                if (!Array.isArray(object.entity_dependencies))
                    throw TypeError(".repository.FullTask.entity_dependencies: array expected");
                message.entity_dependencies = [];
                for (let i = 0; i < object.entity_dependencies.length; ++i)
                    message.entity_dependencies[i] = String(object.entity_dependencies[i]);
            }
            if (object.entity_dependencies_raw != null)
                message.entity_dependencies_raw = String(object.entity_dependencies_raw);
            if (object.dependencies) {
                if (!Array.isArray(object.dependencies))
                    throw TypeError(".repository.FullTask.dependencies: array expected");
                message.dependencies = [];
                for (let i = 0; i < object.dependencies.length; ++i)
                    message.dependencies[i] = String(object.dependencies[i]);
            }
            if (object.dependencies_raw != null)
                message.dependencies_raw = String(object.dependencies_raw);
            if (object.file_status != null)
                message.file_status = String(object.file_status);
            if (object.status != null) {
                if (typeof object.status !== "object")
                    throw TypeError(".repository.FullTask.status: object expected");
                message.status = $root.repository.Status.fromObject(object.status);
            }
            if (object.is_link != null)
                message.is_link = Boolean(object.is_link);
            if (object.pointer != null)
                message.pointer = String(object.pointer);
            if (object.preview_id != null)
                message.preview_id = String(object.preview_id);
            if (object.preview != null)
                if (typeof object.preview === "string")
                    $util.base64.decode(object.preview, message.preview = $util.newBuffer($util.base64.length(object.preview)), 0);
                else if (object.preview.length >= 0)
                    message.preview = object.preview;
            if (object.preview_extension != null)
                message.preview_extension = String(object.preview_extension);
            if (object.checkpoints) {
                if (!Array.isArray(object.checkpoints))
                    throw TypeError(".repository.FullTask.checkpoints: array expected");
                message.checkpoints = [];
                for (let i = 0; i < object.checkpoints.length; ++i) {
                    if (typeof object.checkpoints[i] !== "object")
                        throw TypeError(".repository.FullTask.checkpoints: object expected");
                    message.checkpoints[i] = $root.repository.Checkpoint.fromObject(object.checkpoints[i]);
                }
            }
            if (object.trashed != null)
                message.trashed = Boolean(object.trashed);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            if (object.type != null)
                message.type = String(object.type);
            return message;
        };

        /**
         * Creates a plain object from a FullTask message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.FullTask
         * @static
         * @param {repository.FullTask} message FullTask
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        FullTask.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults) {
                object.tags = [];
                object.entity_dependencies = [];
                object.dependencies = [];
                object.checkpoints = [];
            }
            if (options.defaults) {
                object.id = "";
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.mtime = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.mtime = options.longs === String ? "0" : 0;
                object.created_at = "";
                object.name = "";
                object.description = "";
                object.extension = "";
                object.is_resource = false;
                object.status_id = "";
                object.status_short_name = "";
                object.task_type_id = "";
                object.task_type_name = "";
                object.task_type_icon = "";
                object.entity_id = "";
                object.entity_name = "";
                object.entity_path = "";
                object.task_path = "";
                object.assignee_id = "";
                object.assignee_email = "";
                object.assignee_name = "";
                object.assigner_id = "";
                object.assigner_email = "";
                object.assigner_name = "";
                object.is_dependency = false;
                object.dependency_level = 0;
                object.file_path = "";
                object.tags_raw = "";
                object.entity_dependencies_raw = "";
                object.dependencies_raw = "";
                object.file_status = "";
                object.status = null;
                object.is_link = false;
                object.pointer = "";
                object.preview_id = "";
                if (options.bytes === String)
                    object.preview = "";
                else {
                    object.preview = [];
                    if (options.bytes !== Array)
                        object.preview = $util.newBuffer(object.preview);
                }
                object.preview_extension = "";
                object.trashed = false;
                object.synced = false;
                object.type = "";
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.created_at != null && message.hasOwnProperty("created_at"))
                object.created_at = message.created_at;
            if (message.name != null && message.hasOwnProperty("name"))
                object.name = message.name;
            if (message.description != null && message.hasOwnProperty("description"))
                object.description = message.description;
            if (message.extension != null && message.hasOwnProperty("extension"))
                object.extension = message.extension;
            if (message.is_resource != null && message.hasOwnProperty("is_resource"))
                object.is_resource = message.is_resource;
            if (message.status_id != null && message.hasOwnProperty("status_id"))
                object.status_id = message.status_id;
            if (message.status_short_name != null && message.hasOwnProperty("status_short_name"))
                object.status_short_name = message.status_short_name;
            if (message.task_type_id != null && message.hasOwnProperty("task_type_id"))
                object.task_type_id = message.task_type_id;
            if (message.task_type_name != null && message.hasOwnProperty("task_type_name"))
                object.task_type_name = message.task_type_name;
            if (message.task_type_icon != null && message.hasOwnProperty("task_type_icon"))
                object.task_type_icon = message.task_type_icon;
            if (message.entity_id != null && message.hasOwnProperty("entity_id"))
                object.entity_id = message.entity_id;
            if (message.entity_name != null && message.hasOwnProperty("entity_name"))
                object.entity_name = message.entity_name;
            if (message.entity_path != null && message.hasOwnProperty("entity_path"))
                object.entity_path = message.entity_path;
            if (message.task_path != null && message.hasOwnProperty("task_path"))
                object.task_path = message.task_path;
            if (message.assignee_id != null && message.hasOwnProperty("assignee_id"))
                object.assignee_id = message.assignee_id;
            if (message.assignee_email != null && message.hasOwnProperty("assignee_email"))
                object.assignee_email = message.assignee_email;
            if (message.assignee_name != null && message.hasOwnProperty("assignee_name"))
                object.assignee_name = message.assignee_name;
            if (message.assigner_id != null && message.hasOwnProperty("assigner_id"))
                object.assigner_id = message.assigner_id;
            if (message.assigner_email != null && message.hasOwnProperty("assigner_email"))
                object.assigner_email = message.assigner_email;
            if (message.assigner_name != null && message.hasOwnProperty("assigner_name"))
                object.assigner_name = message.assigner_name;
            if (message.is_dependency != null && message.hasOwnProperty("is_dependency"))
                object.is_dependency = message.is_dependency;
            if (message.dependency_level != null && message.hasOwnProperty("dependency_level"))
                object.dependency_level = message.dependency_level;
            if (message.file_path != null && message.hasOwnProperty("file_path"))
                object.file_path = message.file_path;
            if (message.tags && message.tags.length) {
                object.tags = [];
                for (let j = 0; j < message.tags.length; ++j)
                    object.tags[j] = message.tags[j];
            }
            if (message.tags_raw != null && message.hasOwnProperty("tags_raw"))
                object.tags_raw = message.tags_raw;
            if (message.entity_dependencies && message.entity_dependencies.length) {
                object.entity_dependencies = [];
                for (let j = 0; j < message.entity_dependencies.length; ++j)
                    object.entity_dependencies[j] = message.entity_dependencies[j];
            }
            if (message.entity_dependencies_raw != null && message.hasOwnProperty("entity_dependencies_raw"))
                object.entity_dependencies_raw = message.entity_dependencies_raw;
            if (message.dependencies && message.dependencies.length) {
                object.dependencies = [];
                for (let j = 0; j < message.dependencies.length; ++j)
                    object.dependencies[j] = message.dependencies[j];
            }
            if (message.dependencies_raw != null && message.hasOwnProperty("dependencies_raw"))
                object.dependencies_raw = message.dependencies_raw;
            if (message.file_status != null && message.hasOwnProperty("file_status"))
                object.file_status = message.file_status;
            if (message.status != null && message.hasOwnProperty("status"))
                object.status = $root.repository.Status.toObject(message.status, options);
            if (message.is_link != null && message.hasOwnProperty("is_link"))
                object.is_link = message.is_link;
            if (message.pointer != null && message.hasOwnProperty("pointer"))
                object.pointer = message.pointer;
            if (message.preview_id != null && message.hasOwnProperty("preview_id"))
                object.preview_id = message.preview_id;
            if (message.preview != null && message.hasOwnProperty("preview"))
                object.preview = options.bytes === String ? $util.base64.encode(message.preview, 0, message.preview.length) : options.bytes === Array ? Array.prototype.slice.call(message.preview) : message.preview;
            if (message.preview_extension != null && message.hasOwnProperty("preview_extension"))
                object.preview_extension = message.preview_extension;
            if (message.checkpoints && message.checkpoints.length) {
                object.checkpoints = [];
                for (let j = 0; j < message.checkpoints.length; ++j)
                    object.checkpoints[j] = $root.repository.Checkpoint.toObject(message.checkpoints[j], options);
            }
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                object.trashed = message.trashed;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            if (message.type != null && message.hasOwnProperty("type"))
                object.type = message.type;
            return object;
        };

        /**
         * Converts this FullTask to JSON.
         * @function toJSON
         * @memberof repository.FullTask
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        FullTask.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for FullTask
         * @function getTypeUrl
         * @memberof repository.FullTask
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        FullTask.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.FullTask";
        };

        return FullTask;
    })();

    repository.ChunkInfo = (function() {

        /**
         * Properties of a ChunkInfo.
         * @memberof repository
         * @interface IChunkInfo
         * @property {string|null} [hash] ChunkInfo hash
         * @property {number|Long|null} [size] ChunkInfo size
         */

        /**
         * Constructs a new ChunkInfo.
         * @memberof repository
         * @classdesc Represents a ChunkInfo.
         * @implements IChunkInfo
         * @constructor
         * @param {repository.IChunkInfo=} [properties] Properties to set
         */
        function ChunkInfo(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ChunkInfo hash.
         * @member {string} hash
         * @memberof repository.ChunkInfo
         * @instance
         */
        ChunkInfo.prototype.hash = "";

        /**
         * ChunkInfo size.
         * @member {number|Long} size
         * @memberof repository.ChunkInfo
         * @instance
         */
        ChunkInfo.prototype.size = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Creates a new ChunkInfo instance using the specified properties.
         * @function create
         * @memberof repository.ChunkInfo
         * @static
         * @param {repository.IChunkInfo=} [properties] Properties to set
         * @returns {repository.ChunkInfo} ChunkInfo instance
         */
        ChunkInfo.create = function create(properties) {
            return new ChunkInfo(properties);
        };

        /**
         * Encodes the specified ChunkInfo message. Does not implicitly {@link repository.ChunkInfo.verify|verify} messages.
         * @function encode
         * @memberof repository.ChunkInfo
         * @static
         * @param {repository.IChunkInfo} message ChunkInfo message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ChunkInfo.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.size != null && Object.hasOwnProperty.call(message, "size"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.size);
            if (message.hash != null && Object.hasOwnProperty.call(message, "hash"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.hash);
            return writer;
        };

        /**
         * Encodes the specified ChunkInfo message, length delimited. Does not implicitly {@link repository.ChunkInfo.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.ChunkInfo
         * @static
         * @param {repository.IChunkInfo} message ChunkInfo message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ChunkInfo.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a ChunkInfo message from the specified reader or buffer.
         * @function decode
         * @memberof repository.ChunkInfo
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.ChunkInfo} ChunkInfo
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ChunkInfo.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.ChunkInfo();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 3: {
                        message.hash = reader.string();
                        break;
                    }
                case 2: {
                        message.size = reader.int64();
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a ChunkInfo message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.ChunkInfo
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.ChunkInfo} ChunkInfo
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ChunkInfo.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a ChunkInfo message.
         * @function verify
         * @memberof repository.ChunkInfo
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        ChunkInfo.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.hash != null && message.hasOwnProperty("hash"))
                if (!$util.isString(message.hash))
                    return "hash: string expected";
            if (message.size != null && message.hasOwnProperty("size"))
                if (!$util.isInteger(message.size) && !(message.size && $util.isInteger(message.size.low) && $util.isInteger(message.size.high)))
                    return "size: integer|Long expected";
            return null;
        };

        /**
         * Creates a ChunkInfo message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.ChunkInfo
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.ChunkInfo} ChunkInfo
         */
        ChunkInfo.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.ChunkInfo)
                return object;
            let message = new $root.repository.ChunkInfo();
            if (object.hash != null)
                message.hash = String(object.hash);
            if (object.size != null)
                if ($util.Long)
                    (message.size = $util.Long.fromValue(object.size)).unsigned = false;
                else if (typeof object.size === "string")
                    message.size = parseInt(object.size, 10);
                else if (typeof object.size === "number")
                    message.size = object.size;
                else if (typeof object.size === "object")
                    message.size = new $util.LongBits(object.size.low >>> 0, object.size.high >>> 0).toNumber();
            return message;
        };

        /**
         * Creates a plain object from a ChunkInfo message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.ChunkInfo
         * @static
         * @param {repository.ChunkInfo} message ChunkInfo
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        ChunkInfo.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.defaults) {
                if ($util.Long) {
                    let long = new $util.Long(0, 0, false);
                    object.size = options.longs === String ? long.toString() : options.longs === Number ? long.toNumber() : long;
                } else
                    object.size = options.longs === String ? "0" : 0;
                object.hash = "";
            }
            if (message.size != null && message.hasOwnProperty("size"))
                if (typeof message.size === "number")
                    object.size = options.longs === String ? String(message.size) : message.size;
                else
                    object.size = options.longs === String ? $util.Long.prototype.toString.call(message.size) : options.longs === Number ? new $util.LongBits(message.size.low >>> 0, message.size.high >>> 0).toNumber() : message.size;
            if (message.hash != null && message.hasOwnProperty("hash"))
                object.hash = message.hash;
            return object;
        };

        /**
         * Converts this ChunkInfo to JSON.
         * @function toJSON
         * @memberof repository.ChunkInfo
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        ChunkInfo.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for ChunkInfo
         * @function getTypeUrl
         * @memberof repository.ChunkInfo
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ChunkInfo.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.ChunkInfo";
        };

        return ChunkInfo;
    })();

    repository.FullTaskList = (function() {

        /**
         * Properties of a FullTaskList.
         * @memberof repository
         * @interface IFullTaskList
         * @property {Array.<repository.IFullTask>|null} [full_tasks] FullTaskList full_tasks
         */

        /**
         * Constructs a new FullTaskList.
         * @memberof repository
         * @classdesc Represents a FullTaskList.
         * @implements IFullTaskList
         * @constructor
         * @param {repository.IFullTaskList=} [properties] Properties to set
         */
        function FullTaskList(properties) {
            this.full_tasks = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * FullTaskList full_tasks.
         * @member {Array.<repository.IFullTask>} full_tasks
         * @memberof repository.FullTaskList
         * @instance
         */
        FullTaskList.prototype.full_tasks = $util.emptyArray;

        /**
         * Creates a new FullTaskList instance using the specified properties.
         * @function create
         * @memberof repository.FullTaskList
         * @static
         * @param {repository.IFullTaskList=} [properties] Properties to set
         * @returns {repository.FullTaskList} FullTaskList instance
         */
        FullTaskList.create = function create(properties) {
            return new FullTaskList(properties);
        };

        /**
         * Encodes the specified FullTaskList message. Does not implicitly {@link repository.FullTaskList.verify|verify} messages.
         * @function encode
         * @memberof repository.FullTaskList
         * @static
         * @param {repository.IFullTaskList} message FullTaskList message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FullTaskList.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.full_tasks != null && message.full_tasks.length)
                for (let i = 0; i < message.full_tasks.length; ++i)
                    $root.repository.FullTask.encode(message.full_tasks[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified FullTaskList message, length delimited. Does not implicitly {@link repository.FullTaskList.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.FullTaskList
         * @static
         * @param {repository.IFullTaskList} message FullTaskList message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FullTaskList.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a FullTaskList message from the specified reader or buffer.
         * @function decode
         * @memberof repository.FullTaskList
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.FullTaskList} FullTaskList
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FullTaskList.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.FullTaskList();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        if (!(message.full_tasks && message.full_tasks.length))
                            message.full_tasks = [];
                        message.full_tasks.push($root.repository.FullTask.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a FullTaskList message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.FullTaskList
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.FullTaskList} FullTaskList
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FullTaskList.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a FullTaskList message.
         * @function verify
         * @memberof repository.FullTaskList
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        FullTaskList.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.full_tasks != null && message.hasOwnProperty("full_tasks")) {
                if (!Array.isArray(message.full_tasks))
                    return "full_tasks: array expected";
                for (let i = 0; i < message.full_tasks.length; ++i) {
                    let error = $root.repository.FullTask.verify(message.full_tasks[i]);
                    if (error)
                        return "full_tasks." + error;
                }
            }
            return null;
        };

        /**
         * Creates a FullTaskList message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.FullTaskList
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.FullTaskList} FullTaskList
         */
        FullTaskList.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.FullTaskList)
                return object;
            let message = new $root.repository.FullTaskList();
            if (object.full_tasks) {
                if (!Array.isArray(object.full_tasks))
                    throw TypeError(".repository.FullTaskList.full_tasks: array expected");
                message.full_tasks = [];
                for (let i = 0; i < object.full_tasks.length; ++i) {
                    if (typeof object.full_tasks[i] !== "object")
                        throw TypeError(".repository.FullTaskList.full_tasks: object expected");
                    message.full_tasks[i] = $root.repository.FullTask.fromObject(object.full_tasks[i]);
                }
            }
            return message;
        };

        /**
         * Creates a plain object from a FullTaskList message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.FullTaskList
         * @static
         * @param {repository.FullTaskList} message FullTaskList
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        FullTaskList.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults)
                object.full_tasks = [];
            if (message.full_tasks && message.full_tasks.length) {
                object.full_tasks = [];
                for (let j = 0; j < message.full_tasks.length; ++j)
                    object.full_tasks[j] = $root.repository.FullTask.toObject(message.full_tasks[j], options);
            }
            return object;
        };

        /**
         * Converts this FullTaskList to JSON.
         * @function toJSON
         * @memberof repository.FullTaskList
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        FullTaskList.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for FullTaskList
         * @function getTypeUrl
         * @memberof repository.FullTaskList
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        FullTaskList.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.FullTaskList";
        };

        return FullTaskList;
    })();

    repository.Previews = (function() {

        /**
         * Properties of a Previews.
         * @memberof repository
         * @interface IPreviews
         * @property {Array.<repository.IPreview>|null} [previews] Previews previews
         */

        /**
         * Constructs a new Previews.
         * @memberof repository
         * @classdesc Represents a Previews.
         * @implements IPreviews
         * @constructor
         * @param {repository.IPreviews=} [properties] Properties to set
         */
        function Previews(properties) {
            this.previews = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Previews previews.
         * @member {Array.<repository.IPreview>} previews
         * @memberof repository.Previews
         * @instance
         */
        Previews.prototype.previews = $util.emptyArray;

        /**
         * Creates a new Previews instance using the specified properties.
         * @function create
         * @memberof repository.Previews
         * @static
         * @param {repository.IPreviews=} [properties] Properties to set
         * @returns {repository.Previews} Previews instance
         */
        Previews.create = function create(properties) {
            return new Previews(properties);
        };

        /**
         * Encodes the specified Previews message. Does not implicitly {@link repository.Previews.verify|verify} messages.
         * @function encode
         * @memberof repository.Previews
         * @static
         * @param {repository.IPreviews} message Previews message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Previews.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.previews != null && message.previews.length)
                for (let i = 0; i < message.previews.length; ++i)
                    $root.repository.Preview.encode(message.previews[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified Previews message, length delimited. Does not implicitly {@link repository.Previews.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Previews
         * @static
         * @param {repository.IPreviews} message Previews message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Previews.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Previews message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Previews
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Previews} Previews
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Previews.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Previews();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        if (!(message.previews && message.previews.length))
                            message.previews = [];
                        message.previews.push($root.repository.Preview.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a Previews message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Previews
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Previews} Previews
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Previews.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Previews message.
         * @function verify
         * @memberof repository.Previews
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Previews.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.previews != null && message.hasOwnProperty("previews")) {
                if (!Array.isArray(message.previews))
                    return "previews: array expected";
                for (let i = 0; i < message.previews.length; ++i) {
                    let error = $root.repository.Preview.verify(message.previews[i]);
                    if (error)
                        return "previews." + error;
                }
            }
            return null;
        };

        /**
         * Creates a Previews message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Previews
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Previews} Previews
         */
        Previews.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Previews)
                return object;
            let message = new $root.repository.Previews();
            if (object.previews) {
                if (!Array.isArray(object.previews))
                    throw TypeError(".repository.Previews.previews: array expected");
                message.previews = [];
                for (let i = 0; i < object.previews.length; ++i) {
                    if (typeof object.previews[i] !== "object")
                        throw TypeError(".repository.Previews.previews: object expected");
                    message.previews[i] = $root.repository.Preview.fromObject(object.previews[i]);
                }
            }
            return message;
        };

        /**
         * Creates a plain object from a Previews message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Previews
         * @static
         * @param {repository.Previews} message Previews
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Previews.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults)
                object.previews = [];
            if (message.previews && message.previews.length) {
                object.previews = [];
                for (let j = 0; j < message.previews.length; ++j)
                    object.previews[j] = $root.repository.Preview.toObject(message.previews[j], options);
            }
            return object;
        };

        /**
         * Converts this Previews to JSON.
         * @function toJSON
         * @memberof repository.Previews
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Previews.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Previews
         * @function getTypeUrl
         * @memberof repository.Previews
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Previews.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Previews";
        };

        return Previews;
    })();

    repository.ChunkHashes = (function() {

        /**
         * Properties of a ChunkHashes.
         * @memberof repository
         * @interface IChunkHashes
         * @property {Array.<string>|null} [chunk_hashes] ChunkHashes chunk_hashes
         */

        /**
         * Constructs a new ChunkHashes.
         * @memberof repository
         * @classdesc Represents a ChunkHashes.
         * @implements IChunkHashes
         * @constructor
         * @param {repository.IChunkHashes=} [properties] Properties to set
         */
        function ChunkHashes(properties) {
            this.chunk_hashes = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ChunkHashes chunk_hashes.
         * @member {Array.<string>} chunk_hashes
         * @memberof repository.ChunkHashes
         * @instance
         */
        ChunkHashes.prototype.chunk_hashes = $util.emptyArray;

        /**
         * Creates a new ChunkHashes instance using the specified properties.
         * @function create
         * @memberof repository.ChunkHashes
         * @static
         * @param {repository.IChunkHashes=} [properties] Properties to set
         * @returns {repository.ChunkHashes} ChunkHashes instance
         */
        ChunkHashes.create = function create(properties) {
            return new ChunkHashes(properties);
        };

        /**
         * Encodes the specified ChunkHashes message. Does not implicitly {@link repository.ChunkHashes.verify|verify} messages.
         * @function encode
         * @memberof repository.ChunkHashes
         * @static
         * @param {repository.IChunkHashes} message ChunkHashes message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ChunkHashes.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.chunk_hashes != null && message.chunk_hashes.length)
                for (let i = 0; i < message.chunk_hashes.length; ++i)
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.chunk_hashes[i]);
            return writer;
        };

        /**
         * Encodes the specified ChunkHashes message, length delimited. Does not implicitly {@link repository.ChunkHashes.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.ChunkHashes
         * @static
         * @param {repository.IChunkHashes} message ChunkHashes message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ChunkHashes.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a ChunkHashes message from the specified reader or buffer.
         * @function decode
         * @memberof repository.ChunkHashes
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.ChunkHashes} ChunkHashes
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ChunkHashes.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.ChunkHashes();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        if (!(message.chunk_hashes && message.chunk_hashes.length))
                            message.chunk_hashes = [];
                        message.chunk_hashes.push(reader.string());
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a ChunkHashes message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.ChunkHashes
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.ChunkHashes} ChunkHashes
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ChunkHashes.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a ChunkHashes message.
         * @function verify
         * @memberof repository.ChunkHashes
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        ChunkHashes.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.chunk_hashes != null && message.hasOwnProperty("chunk_hashes")) {
                if (!Array.isArray(message.chunk_hashes))
                    return "chunk_hashes: array expected";
                for (let i = 0; i < message.chunk_hashes.length; ++i)
                    if (!$util.isString(message.chunk_hashes[i]))
                        return "chunk_hashes: string[] expected";
            }
            return null;
        };

        /**
         * Creates a ChunkHashes message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.ChunkHashes
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.ChunkHashes} ChunkHashes
         */
        ChunkHashes.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.ChunkHashes)
                return object;
            let message = new $root.repository.ChunkHashes();
            if (object.chunk_hashes) {
                if (!Array.isArray(object.chunk_hashes))
                    throw TypeError(".repository.ChunkHashes.chunk_hashes: array expected");
                message.chunk_hashes = [];
                for (let i = 0; i < object.chunk_hashes.length; ++i)
                    message.chunk_hashes[i] = String(object.chunk_hashes[i]);
            }
            return message;
        };

        /**
         * Creates a plain object from a ChunkHashes message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.ChunkHashes
         * @static
         * @param {repository.ChunkHashes} message ChunkHashes
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        ChunkHashes.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults)
                object.chunk_hashes = [];
            if (message.chunk_hashes && message.chunk_hashes.length) {
                object.chunk_hashes = [];
                for (let j = 0; j < message.chunk_hashes.length; ++j)
                    object.chunk_hashes[j] = message.chunk_hashes[j];
            }
            return object;
        };

        /**
         * Converts this ChunkHashes to JSON.
         * @function toJSON
         * @memberof repository.ChunkHashes
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        ChunkHashes.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for ChunkHashes
         * @function getTypeUrl
         * @memberof repository.ChunkHashes
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ChunkHashes.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.ChunkHashes";
        };

        return ChunkHashes;
    })();

    repository.ChunkInfos = (function() {

        /**
         * Properties of a ChunkInfos.
         * @memberof repository
         * @interface IChunkInfos
         * @property {Array.<repository.IChunkInfo>|null} [chunk_infos] ChunkInfos chunk_infos
         */

        /**
         * Constructs a new ChunkInfos.
         * @memberof repository
         * @classdesc Represents a ChunkInfos.
         * @implements IChunkInfos
         * @constructor
         * @param {repository.IChunkInfos=} [properties] Properties to set
         */
        function ChunkInfos(properties) {
            this.chunk_infos = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ChunkInfos chunk_infos.
         * @member {Array.<repository.IChunkInfo>} chunk_infos
         * @memberof repository.ChunkInfos
         * @instance
         */
        ChunkInfos.prototype.chunk_infos = $util.emptyArray;

        /**
         * Creates a new ChunkInfos instance using the specified properties.
         * @function create
         * @memberof repository.ChunkInfos
         * @static
         * @param {repository.IChunkInfos=} [properties] Properties to set
         * @returns {repository.ChunkInfos} ChunkInfos instance
         */
        ChunkInfos.create = function create(properties) {
            return new ChunkInfos(properties);
        };

        /**
         * Encodes the specified ChunkInfos message. Does not implicitly {@link repository.ChunkInfos.verify|verify} messages.
         * @function encode
         * @memberof repository.ChunkInfos
         * @static
         * @param {repository.IChunkInfos} message ChunkInfos message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ChunkInfos.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.chunk_infos != null && message.chunk_infos.length)
                for (let i = 0; i < message.chunk_infos.length; ++i)
                    $root.repository.ChunkInfo.encode(message.chunk_infos[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified ChunkInfos message, length delimited. Does not implicitly {@link repository.ChunkInfos.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.ChunkInfos
         * @static
         * @param {repository.IChunkInfos} message ChunkInfos message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ChunkInfos.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a ChunkInfos message from the specified reader or buffer.
         * @function decode
         * @memberof repository.ChunkInfos
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.ChunkInfos} ChunkInfos
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ChunkInfos.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.ChunkInfos();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        if (!(message.chunk_infos && message.chunk_infos.length))
                            message.chunk_infos = [];
                        message.chunk_infos.push($root.repository.ChunkInfo.decode(reader, reader.uint32()));
                        break;
                    }
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes a ChunkInfos message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.ChunkInfos
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.ChunkInfos} ChunkInfos
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ChunkInfos.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a ChunkInfos message.
         * @function verify
         * @memberof repository.ChunkInfos
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        ChunkInfos.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.chunk_infos != null && message.hasOwnProperty("chunk_infos")) {
                if (!Array.isArray(message.chunk_infos))
                    return "chunk_infos: array expected";
                for (let i = 0; i < message.chunk_infos.length; ++i) {
                    let error = $root.repository.ChunkInfo.verify(message.chunk_infos[i]);
                    if (error)
                        return "chunk_infos." + error;
                }
            }
            return null;
        };

        /**
         * Creates a ChunkInfos message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.ChunkInfos
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.ChunkInfos} ChunkInfos
         */
        ChunkInfos.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.ChunkInfos)
                return object;
            let message = new $root.repository.ChunkInfos();
            if (object.chunk_infos) {
                if (!Array.isArray(object.chunk_infos))
                    throw TypeError(".repository.ChunkInfos.chunk_infos: array expected");
                message.chunk_infos = [];
                for (let i = 0; i < object.chunk_infos.length; ++i) {
                    if (typeof object.chunk_infos[i] !== "object")
                        throw TypeError(".repository.ChunkInfos.chunk_infos: object expected");
                    message.chunk_infos[i] = $root.repository.ChunkInfo.fromObject(object.chunk_infos[i]);
                }
            }
            return message;
        };

        /**
         * Creates a plain object from a ChunkInfos message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.ChunkInfos
         * @static
         * @param {repository.ChunkInfos} message ChunkInfos
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        ChunkInfos.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults)
                object.chunk_infos = [];
            if (message.chunk_infos && message.chunk_infos.length) {
                object.chunk_infos = [];
                for (let j = 0; j < message.chunk_infos.length; ++j)
                    object.chunk_infos[j] = $root.repository.ChunkInfo.toObject(message.chunk_infos[j], options);
            }
            return object;
        };

        /**
         * Converts this ChunkInfos to JSON.
         * @function toJSON
         * @memberof repository.ChunkInfos
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        ChunkInfos.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for ChunkInfos
         * @function getTypeUrl
         * @memberof repository.ChunkInfos
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        ChunkInfos.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.ChunkInfos";
        };

        return ChunkInfos;
    })();

    return repository;
})();

export { $root as default };
