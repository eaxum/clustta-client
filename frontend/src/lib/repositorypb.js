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

    repository.CollectionType = (function() {

        /**
         * Properties of a CollectionType.
         * @memberof repository
         * @interface ICollectionType
         * @property {string|null} [id] CollectionType id
         * @property {number|Long|null} [mtime] CollectionType mtime
         * @property {string|null} [name] CollectionType name
         * @property {string|null} [icon] CollectionType icon
         * @property {boolean|null} [synced] CollectionType synced
         */

        /**
         * Constructs a new CollectionType.
         * @memberof repository
         * @classdesc Represents a CollectionType.
         * @implements ICollectionType
         * @constructor
         * @param {repository.ICollectionType=} [properties] Properties to set
         */
        function CollectionType(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CollectionType id.
         * @member {string} id
         * @memberof repository.CollectionType
         * @instance
         */
        CollectionType.prototype.id = "";

        /**
         * CollectionType mtime.
         * @member {number|Long} mtime
         * @memberof repository.CollectionType
         * @instance
         */
        CollectionType.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CollectionType name.
         * @member {string} name
         * @memberof repository.CollectionType
         * @instance
         */
        CollectionType.prototype.name = "";

        /**
         * CollectionType icon.
         * @member {string} icon
         * @memberof repository.CollectionType
         * @instance
         */
        CollectionType.prototype.icon = "";

        /**
         * CollectionType synced.
         * @member {boolean} synced
         * @memberof repository.CollectionType
         * @instance
         */
        CollectionType.prototype.synced = false;

        /**
         * Creates a new CollectionType instance using the specified properties.
         * @function create
         * @memberof repository.CollectionType
         * @static
         * @param {repository.ICollectionType=} [properties] Properties to set
         * @returns {repository.CollectionType} CollectionType instance
         */
        CollectionType.create = function create(properties) {
            return new CollectionType(properties);
        };

        /**
         * Encodes the specified CollectionType message. Does not implicitly {@link repository.CollectionType.verify|verify} messages.
         * @function encode
         * @memberof repository.CollectionType
         * @static
         * @param {repository.ICollectionType} message CollectionType message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CollectionType.encode = function encode(message, writer) {
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
         * Encodes the specified CollectionType message, length delimited. Does not implicitly {@link repository.CollectionType.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.CollectionType
         * @static
         * @param {repository.ICollectionType} message CollectionType message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CollectionType.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a CollectionType message from the specified reader or buffer.
         * @function decode
         * @memberof repository.CollectionType
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.CollectionType} CollectionType
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CollectionType.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.CollectionType();
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
         * Decodes a CollectionType message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.CollectionType
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.CollectionType} CollectionType
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CollectionType.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a CollectionType message.
         * @function verify
         * @memberof repository.CollectionType
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        CollectionType.verify = function verify(message) {
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
         * Creates a CollectionType message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.CollectionType
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.CollectionType} CollectionType
         */
        CollectionType.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.CollectionType)
                return object;
            let message = new $root.repository.CollectionType();
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
         * Creates a plain object from a CollectionType message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.CollectionType
         * @static
         * @param {repository.CollectionType} message CollectionType
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        CollectionType.toObject = function toObject(message, options) {
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
         * Converts this CollectionType to JSON.
         * @function toJSON
         * @memberof repository.CollectionType
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        CollectionType.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for CollectionType
         * @function getTypeUrl
         * @memberof repository.CollectionType
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CollectionType.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.CollectionType";
        };

        return CollectionType;
    })();

    repository.AssetType = (function() {

        /**
         * Properties of an AssetType.
         * @memberof repository
         * @interface IAssetType
         * @property {string|null} [id] AssetType id
         * @property {number|Long|null} [mtime] AssetType mtime
         * @property {string|null} [name] AssetType name
         * @property {string|null} [icon] AssetType icon
         * @property {boolean|null} [synced] AssetType synced
         */

        /**
         * Constructs a new AssetType.
         * @memberof repository
         * @classdesc Represents an AssetType.
         * @implements IAssetType
         * @constructor
         * @param {repository.IAssetType=} [properties] Properties to set
         */
        function AssetType(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * AssetType id.
         * @member {string} id
         * @memberof repository.AssetType
         * @instance
         */
        AssetType.prototype.id = "";

        /**
         * AssetType mtime.
         * @member {number|Long} mtime
         * @memberof repository.AssetType
         * @instance
         */
        AssetType.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * AssetType name.
         * @member {string} name
         * @memberof repository.AssetType
         * @instance
         */
        AssetType.prototype.name = "";

        /**
         * AssetType icon.
         * @member {string} icon
         * @memberof repository.AssetType
         * @instance
         */
        AssetType.prototype.icon = "";

        /**
         * AssetType synced.
         * @member {boolean} synced
         * @memberof repository.AssetType
         * @instance
         */
        AssetType.prototype.synced = false;

        /**
         * Creates a new AssetType instance using the specified properties.
         * @function create
         * @memberof repository.AssetType
         * @static
         * @param {repository.IAssetType=} [properties] Properties to set
         * @returns {repository.AssetType} AssetType instance
         */
        AssetType.create = function create(properties) {
            return new AssetType(properties);
        };

        /**
         * Encodes the specified AssetType message. Does not implicitly {@link repository.AssetType.verify|verify} messages.
         * @function encode
         * @memberof repository.AssetType
         * @static
         * @param {repository.IAssetType} message AssetType message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AssetType.encode = function encode(message, writer) {
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
         * Encodes the specified AssetType message, length delimited. Does not implicitly {@link repository.AssetType.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.AssetType
         * @static
         * @param {repository.IAssetType} message AssetType message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AssetType.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an AssetType message from the specified reader or buffer.
         * @function decode
         * @memberof repository.AssetType
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.AssetType} AssetType
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AssetType.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.AssetType();
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
         * Decodes an AssetType message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.AssetType
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.AssetType} AssetType
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AssetType.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an AssetType message.
         * @function verify
         * @memberof repository.AssetType
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        AssetType.verify = function verify(message) {
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
         * Creates an AssetType message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.AssetType
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.AssetType} AssetType
         */
        AssetType.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.AssetType)
                return object;
            let message = new $root.repository.AssetType();
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
         * Creates a plain object from an AssetType message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.AssetType
         * @static
         * @param {repository.AssetType} message AssetType
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        AssetType.toObject = function toObject(message, options) {
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
         * Converts this AssetType to JSON.
         * @function toJSON
         * @memberof repository.AssetType
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        AssetType.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for AssetType
         * @function getTypeUrl
         * @memberof repository.AssetType
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AssetType.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.AssetType";
        };

        return AssetType;
    })();

    repository.Asset = (function() {

        /**
         * Properties of an Asset.
         * @memberof repository
         * @interface IAsset
         * @property {string|null} [id] Asset id
         * @property {number|Long|null} [mtime] Asset mtime
         * @property {string|null} [created_at] Asset created_at
         * @property {string|null} [name] Asset name
         * @property {string|null} [description] Asset description
         * @property {string|null} [extension] Asset extension
         * @property {boolean|null} [is_resource] Asset is_resource
         * @property {string|null} [status_id] Asset status_id
         * @property {string|null} [asset_type_id] Asset asset_type_id
         * @property {string|null} [collection_id] Asset collection_id
         * @property {string|null} [assignee_id] Asset assignee_id
         * @property {string|null} [assigner_id] Asset assigner_id
         * @property {boolean|null} [is_link] Asset is_link
         * @property {string|null} [pointer] Asset pointer
         * @property {string|null} [preview_id] Asset preview_id
         * @property {boolean|null} [trashed] Asset trashed
         * @property {boolean|null} [synced] Asset synced
         */

        /**
         * Constructs a new Asset.
         * @memberof repository
         * @classdesc Represents an Asset.
         * @implements IAsset
         * @constructor
         * @param {repository.IAsset=} [properties] Properties to set
         */
        function Asset(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Asset id.
         * @member {string} id
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.id = "";

        /**
         * Asset mtime.
         * @member {number|Long} mtime
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Asset created_at.
         * @member {string} created_at
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.created_at = "";

        /**
         * Asset name.
         * @member {string} name
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.name = "";

        /**
         * Asset description.
         * @member {string} description
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.description = "";

        /**
         * Asset extension.
         * @member {string} extension
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.extension = "";

        /**
         * Asset is_resource.
         * @member {boolean} is_resource
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.is_resource = false;

        /**
         * Asset status_id.
         * @member {string} status_id
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.status_id = "";

        /**
         * Asset asset_type_id.
         * @member {string} asset_type_id
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.asset_type_id = "";

        /**
         * Asset collection_id.
         * @member {string} collection_id
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.collection_id = "";

        /**
         * Asset assignee_id.
         * @member {string} assignee_id
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.assignee_id = "";

        /**
         * Asset assigner_id.
         * @member {string} assigner_id
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.assigner_id = "";

        /**
         * Asset is_link.
         * @member {boolean} is_link
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.is_link = false;

        /**
         * Asset pointer.
         * @member {string} pointer
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.pointer = "";

        /**
         * Asset preview_id.
         * @member {string} preview_id
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.preview_id = "";

        /**
         * Asset trashed.
         * @member {boolean} trashed
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.trashed = false;

        /**
         * Asset synced.
         * @member {boolean} synced
         * @memberof repository.Asset
         * @instance
         */
        Asset.prototype.synced = false;

        /**
         * Creates a new Asset instance using the specified properties.
         * @function create
         * @memberof repository.Asset
         * @static
         * @param {repository.IAsset=} [properties] Properties to set
         * @returns {repository.Asset} Asset instance
         */
        Asset.create = function create(properties) {
            return new Asset(properties);
        };

        /**
         * Encodes the specified Asset message. Does not implicitly {@link repository.Asset.verify|verify} messages.
         * @function encode
         * @memberof repository.Asset
         * @static
         * @param {repository.IAsset} message Asset message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Asset.encode = function encode(message, writer) {
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
            if (message.asset_type_id != null && Object.hasOwnProperty.call(message, "asset_type_id"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.asset_type_id);
            if (message.collection_id != null && Object.hasOwnProperty.call(message, "collection_id"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.collection_id);
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
         * Encodes the specified Asset message, length delimited. Does not implicitly {@link repository.Asset.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Asset
         * @static
         * @param {repository.IAsset} message Asset message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Asset.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an Asset message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Asset
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Asset} Asset
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Asset.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Asset();
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
                        message.asset_type_id = reader.string();
                        break;
                    }
                case 10: {
                        message.collection_id = reader.string();
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
         * Decodes an Asset message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Asset
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Asset} Asset
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Asset.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an Asset message.
         * @function verify
         * @memberof repository.Asset
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Asset.verify = function verify(message) {
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
            if (message.asset_type_id != null && message.hasOwnProperty("asset_type_id"))
                if (!$util.isString(message.asset_type_id))
                    return "asset_type_id: string expected";
            if (message.collection_id != null && message.hasOwnProperty("collection_id"))
                if (!$util.isString(message.collection_id))
                    return "collection_id: string expected";
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
         * Creates an Asset message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Asset
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Asset} Asset
         */
        Asset.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Asset)
                return object;
            let message = new $root.repository.Asset();
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
            if (object.asset_type_id != null)
                message.asset_type_id = String(object.asset_type_id);
            if (object.collection_id != null)
                message.collection_id = String(object.collection_id);
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
         * Creates a plain object from an Asset message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Asset
         * @static
         * @param {repository.Asset} message Asset
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Asset.toObject = function toObject(message, options) {
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
                object.asset_type_id = "";
                object.collection_id = "";
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
            if (message.asset_type_id != null && message.hasOwnProperty("asset_type_id"))
                object.asset_type_id = message.asset_type_id;
            if (message.collection_id != null && message.hasOwnProperty("collection_id"))
                object.collection_id = message.collection_id;
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
         * Converts this Asset to JSON.
         * @function toJSON
         * @memberof repository.Asset
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Asset.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Asset
         * @function getTypeUrl
         * @memberof repository.Asset
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Asset.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Asset";
        };

        return Asset;
    })();

    repository.Collection = (function() {

        /**
         * Properties of a Collection.
         * @memberof repository
         * @interface ICollection
         * @property {string|null} [id] Collection id
         * @property {number|Long|null} [mtime] Collection mtime
         * @property {string|null} [created_at] Collection created_at
         * @property {string|null} [name] Collection name
         * @property {string|null} [collection_path] Collection collection_path
         * @property {string|null} [description] Collection description
         * @property {boolean|null} [trashed] Collection trashed
         * @property {string|null} [collection_type_id] Collection collection_type_id
         * @property {string|null} [parent_id] Collection parent_id
         * @property {string|null} [preview_id] Collection preview_id
         * @property {boolean|null} [synced] Collection synced
         * @property {boolean|null} [is_library] Collection is_library
         */

        /**
         * Constructs a new Collection.
         * @memberof repository
         * @classdesc Represents a Collection.
         * @implements ICollection
         * @constructor
         * @param {repository.ICollection=} [properties] Properties to set
         */
        function Collection(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * Collection id.
         * @member {string} id
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.id = "";

        /**
         * Collection mtime.
         * @member {number|Long} mtime
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * Collection created_at.
         * @member {string} created_at
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.created_at = "";

        /**
         * Collection name.
         * @member {string} name
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.name = "";

        /**
         * Collection collection_path.
         * @member {string} collection_path
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.collection_path = "";

        /**
         * Collection description.
         * @member {string} description
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.description = "";

        /**
         * Collection trashed.
         * @member {boolean} trashed
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.trashed = false;

        /**
         * Collection collection_type_id.
         * @member {string} collection_type_id
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.collection_type_id = "";

        /**
         * Collection parent_id.
         * @member {string} parent_id
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.parent_id = "";

        /**
         * Collection preview_id.
         * @member {string} preview_id
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.preview_id = "";

        /**
         * Collection synced.
         * @member {boolean} synced
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.synced = false;

        /**
         * Collection is_library.
         * @member {boolean} is_library
         * @memberof repository.Collection
         * @instance
         */
        Collection.prototype.is_library = false;

        /**
         * Creates a new Collection instance using the specified properties.
         * @function create
         * @memberof repository.Collection
         * @static
         * @param {repository.ICollection=} [properties] Properties to set
         * @returns {repository.Collection} Collection instance
         */
        Collection.create = function create(properties) {
            return new Collection(properties);
        };

        /**
         * Encodes the specified Collection message. Does not implicitly {@link repository.Collection.verify|verify} messages.
         * @function encode
         * @memberof repository.Collection
         * @static
         * @param {repository.ICollection} message Collection message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Collection.encode = function encode(message, writer) {
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
            if (message.collection_path != null && Object.hasOwnProperty.call(message, "collection_path"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.collection_path);
            if (message.description != null && Object.hasOwnProperty.call(message, "description"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.description);
            if (message.trashed != null && Object.hasOwnProperty.call(message, "trashed"))
                writer.uint32(/* id 7, wireType 0 =*/56).bool(message.trashed);
            if (message.collection_type_id != null && Object.hasOwnProperty.call(message, "collection_type_id"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.collection_type_id);
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
         * Encodes the specified Collection message, length delimited. Does not implicitly {@link repository.Collection.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.Collection
         * @static
         * @param {repository.ICollection} message Collection message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        Collection.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a Collection message from the specified reader or buffer.
         * @function decode
         * @memberof repository.Collection
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.Collection} Collection
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Collection.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.Collection();
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
                        message.collection_path = reader.string();
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
                        message.collection_type_id = reader.string();
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
         * Decodes a Collection message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.Collection
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.Collection} Collection
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        Collection.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a Collection message.
         * @function verify
         * @memberof repository.Collection
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        Collection.verify = function verify(message) {
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
            if (message.collection_path != null && message.hasOwnProperty("collection_path"))
                if (!$util.isString(message.collection_path))
                    return "collection_path: string expected";
            if (message.description != null && message.hasOwnProperty("description"))
                if (!$util.isString(message.description))
                    return "description: string expected";
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                if (typeof message.trashed !== "boolean")
                    return "trashed: boolean expected";
            if (message.collection_type_id != null && message.hasOwnProperty("collection_type_id"))
                if (!$util.isString(message.collection_type_id))
                    return "collection_type_id: string expected";
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
         * Creates a Collection message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.Collection
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.Collection} Collection
         */
        Collection.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.Collection)
                return object;
            let message = new $root.repository.Collection();
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
            if (object.collection_path != null)
                message.collection_path = String(object.collection_path);
            if (object.description != null)
                message.description = String(object.description);
            if (object.trashed != null)
                message.trashed = Boolean(object.trashed);
            if (object.collection_type_id != null)
                message.collection_type_id = String(object.collection_type_id);
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
         * Creates a plain object from a Collection message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.Collection
         * @static
         * @param {repository.Collection} message Collection
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        Collection.toObject = function toObject(message, options) {
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
                object.collection_path = "";
                object.description = "";
                object.trashed = false;
                object.collection_type_id = "";
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
            if (message.collection_path != null && message.hasOwnProperty("collection_path"))
                object.collection_path = message.collection_path;
            if (message.description != null && message.hasOwnProperty("description"))
                object.description = message.description;
            if (message.trashed != null && message.hasOwnProperty("trashed"))
                object.trashed = message.trashed;
            if (message.collection_type_id != null && message.hasOwnProperty("collection_type_id"))
                object.collection_type_id = message.collection_type_id;
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
         * Converts this Collection to JSON.
         * @function toJSON
         * @memberof repository.Collection
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        Collection.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for Collection
         * @function getTypeUrl
         * @memberof repository.Collection
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        Collection.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.Collection";
        };

        return Collection;
    })();

    repository.CollectionAssignee = (function() {

        /**
         * Properties of a CollectionAssignee.
         * @memberof repository
         * @interface ICollectionAssignee
         * @property {string|null} [id] CollectionAssignee id
         * @property {number|Long|null} [mtime] CollectionAssignee mtime
         * @property {string|null} [collection_id] CollectionAssignee collection_id
         * @property {string|null} [assignee_id] CollectionAssignee assignee_id
         * @property {string|null} [assigner_id] CollectionAssignee assigner_id
         * @property {boolean|null} [synced] CollectionAssignee synced
         */

        /**
         * Constructs a new CollectionAssignee.
         * @memberof repository
         * @classdesc Represents a CollectionAssignee.
         * @implements ICollectionAssignee
         * @constructor
         * @param {repository.ICollectionAssignee=} [properties] Properties to set
         */
        function CollectionAssignee(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CollectionAssignee id.
         * @member {string} id
         * @memberof repository.CollectionAssignee
         * @instance
         */
        CollectionAssignee.prototype.id = "";

        /**
         * CollectionAssignee mtime.
         * @member {number|Long} mtime
         * @memberof repository.CollectionAssignee
         * @instance
         */
        CollectionAssignee.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CollectionAssignee collection_id.
         * @member {string} collection_id
         * @memberof repository.CollectionAssignee
         * @instance
         */
        CollectionAssignee.prototype.collection_id = "";

        /**
         * CollectionAssignee assignee_id.
         * @member {string} assignee_id
         * @memberof repository.CollectionAssignee
         * @instance
         */
        CollectionAssignee.prototype.assignee_id = "";

        /**
         * CollectionAssignee assigner_id.
         * @member {string} assigner_id
         * @memberof repository.CollectionAssignee
         * @instance
         */
        CollectionAssignee.prototype.assigner_id = "";

        /**
         * CollectionAssignee synced.
         * @member {boolean} synced
         * @memberof repository.CollectionAssignee
         * @instance
         */
        CollectionAssignee.prototype.synced = false;

        /**
         * Creates a new CollectionAssignee instance using the specified properties.
         * @function create
         * @memberof repository.CollectionAssignee
         * @static
         * @param {repository.ICollectionAssignee=} [properties] Properties to set
         * @returns {repository.CollectionAssignee} CollectionAssignee instance
         */
        CollectionAssignee.create = function create(properties) {
            return new CollectionAssignee(properties);
        };

        /**
         * Encodes the specified CollectionAssignee message. Does not implicitly {@link repository.CollectionAssignee.verify|verify} messages.
         * @function encode
         * @memberof repository.CollectionAssignee
         * @static
         * @param {repository.ICollectionAssignee} message CollectionAssignee message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CollectionAssignee.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.collection_id != null && Object.hasOwnProperty.call(message, "collection_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.collection_id);
            if (message.assignee_id != null && Object.hasOwnProperty.call(message, "assignee_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.assignee_id);
            if (message.assigner_id != null && Object.hasOwnProperty.call(message, "assigner_id"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.assigner_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified CollectionAssignee message, length delimited. Does not implicitly {@link repository.CollectionAssignee.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.CollectionAssignee
         * @static
         * @param {repository.ICollectionAssignee} message CollectionAssignee message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CollectionAssignee.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a CollectionAssignee message from the specified reader or buffer.
         * @function decode
         * @memberof repository.CollectionAssignee
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.CollectionAssignee} CollectionAssignee
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CollectionAssignee.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.CollectionAssignee();
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
                        message.collection_id = reader.string();
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
         * Decodes a CollectionAssignee message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.CollectionAssignee
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.CollectionAssignee} CollectionAssignee
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CollectionAssignee.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a CollectionAssignee message.
         * @function verify
         * @memberof repository.CollectionAssignee
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        CollectionAssignee.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.collection_id != null && message.hasOwnProperty("collection_id"))
                if (!$util.isString(message.collection_id))
                    return "collection_id: string expected";
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
         * Creates a CollectionAssignee message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.CollectionAssignee
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.CollectionAssignee} CollectionAssignee
         */
        CollectionAssignee.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.CollectionAssignee)
                return object;
            let message = new $root.repository.CollectionAssignee();
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
            if (object.collection_id != null)
                message.collection_id = String(object.collection_id);
            if (object.assignee_id != null)
                message.assignee_id = String(object.assignee_id);
            if (object.assigner_id != null)
                message.assigner_id = String(object.assigner_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a CollectionAssignee message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.CollectionAssignee
         * @static
         * @param {repository.CollectionAssignee} message CollectionAssignee
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        CollectionAssignee.toObject = function toObject(message, options) {
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
                object.collection_id = "";
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
            if (message.collection_id != null && message.hasOwnProperty("collection_id"))
                object.collection_id = message.collection_id;
            if (message.assignee_id != null && message.hasOwnProperty("assignee_id"))
                object.assignee_id = message.assignee_id;
            if (message.assigner_id != null && message.hasOwnProperty("assigner_id"))
                object.assigner_id = message.assigner_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this CollectionAssignee to JSON.
         * @function toJSON
         * @memberof repository.CollectionAssignee
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        CollectionAssignee.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for CollectionAssignee
         * @function getTypeUrl
         * @memberof repository.CollectionAssignee
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CollectionAssignee.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.CollectionAssignee";
        };

        return CollectionAssignee;
    })();

    repository.AssetDependency = (function() {

        /**
         * Properties of an AssetDependency.
         * @memberof repository
         * @interface IAssetDependency
         * @property {string|null} [id] AssetDependency id
         * @property {number|Long|null} [mtime] AssetDependency mtime
         * @property {string|null} [asset_id] AssetDependency asset_id
         * @property {string|null} [dependency_id] AssetDependency dependency_id
         * @property {string|null} [dependency_type_id] AssetDependency dependency_type_id
         * @property {boolean|null} [synced] AssetDependency synced
         */

        /**
         * Constructs a new AssetDependency.
         * @memberof repository
         * @classdesc Represents an AssetDependency.
         * @implements IAssetDependency
         * @constructor
         * @param {repository.IAssetDependency=} [properties] Properties to set
         */
        function AssetDependency(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * AssetDependency id.
         * @member {string} id
         * @memberof repository.AssetDependency
         * @instance
         */
        AssetDependency.prototype.id = "";

        /**
         * AssetDependency mtime.
         * @member {number|Long} mtime
         * @memberof repository.AssetDependency
         * @instance
         */
        AssetDependency.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * AssetDependency asset_id.
         * @member {string} asset_id
         * @memberof repository.AssetDependency
         * @instance
         */
        AssetDependency.prototype.asset_id = "";

        /**
         * AssetDependency dependency_id.
         * @member {string} dependency_id
         * @memberof repository.AssetDependency
         * @instance
         */
        AssetDependency.prototype.dependency_id = "";

        /**
         * AssetDependency dependency_type_id.
         * @member {string} dependency_type_id
         * @memberof repository.AssetDependency
         * @instance
         */
        AssetDependency.prototype.dependency_type_id = "";

        /**
         * AssetDependency synced.
         * @member {boolean} synced
         * @memberof repository.AssetDependency
         * @instance
         */
        AssetDependency.prototype.synced = false;

        /**
         * Creates a new AssetDependency instance using the specified properties.
         * @function create
         * @memberof repository.AssetDependency
         * @static
         * @param {repository.IAssetDependency=} [properties] Properties to set
         * @returns {repository.AssetDependency} AssetDependency instance
         */
        AssetDependency.create = function create(properties) {
            return new AssetDependency(properties);
        };

        /**
         * Encodes the specified AssetDependency message. Does not implicitly {@link repository.AssetDependency.verify|verify} messages.
         * @function encode
         * @memberof repository.AssetDependency
         * @static
         * @param {repository.IAssetDependency} message AssetDependency message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AssetDependency.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.asset_id != null && Object.hasOwnProperty.call(message, "asset_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.asset_id);
            if (message.dependency_id != null && Object.hasOwnProperty.call(message, "dependency_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.dependency_id);
            if (message.dependency_type_id != null && Object.hasOwnProperty.call(message, "dependency_type_id"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.dependency_type_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified AssetDependency message, length delimited. Does not implicitly {@link repository.AssetDependency.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.AssetDependency
         * @static
         * @param {repository.IAssetDependency} message AssetDependency message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AssetDependency.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an AssetDependency message from the specified reader or buffer.
         * @function decode
         * @memberof repository.AssetDependency
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.AssetDependency} AssetDependency
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AssetDependency.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.AssetDependency();
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
                        message.asset_id = reader.string();
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
         * Decodes an AssetDependency message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.AssetDependency
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.AssetDependency} AssetDependency
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AssetDependency.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an AssetDependency message.
         * @function verify
         * @memberof repository.AssetDependency
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        AssetDependency.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.asset_id != null && message.hasOwnProperty("asset_id"))
                if (!$util.isString(message.asset_id))
                    return "asset_id: string expected";
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
         * Creates an AssetDependency message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.AssetDependency
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.AssetDependency} AssetDependency
         */
        AssetDependency.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.AssetDependency)
                return object;
            let message = new $root.repository.AssetDependency();
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
            if (object.asset_id != null)
                message.asset_id = String(object.asset_id);
            if (object.dependency_id != null)
                message.dependency_id = String(object.dependency_id);
            if (object.dependency_type_id != null)
                message.dependency_type_id = String(object.dependency_type_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from an AssetDependency message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.AssetDependency
         * @static
         * @param {repository.AssetDependency} message AssetDependency
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        AssetDependency.toObject = function toObject(message, options) {
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
                object.asset_id = "";
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
            if (message.asset_id != null && message.hasOwnProperty("asset_id"))
                object.asset_id = message.asset_id;
            if (message.dependency_id != null && message.hasOwnProperty("dependency_id"))
                object.dependency_id = message.dependency_id;
            if (message.dependency_type_id != null && message.hasOwnProperty("dependency_type_id"))
                object.dependency_type_id = message.dependency_type_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this AssetDependency to JSON.
         * @function toJSON
         * @memberof repository.AssetDependency
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        AssetDependency.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for AssetDependency
         * @function getTypeUrl
         * @memberof repository.AssetDependency
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AssetDependency.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.AssetDependency";
        };

        return AssetDependency;
    })();

    repository.CollectionDependency = (function() {

        /**
         * Properties of a CollectionDependency.
         * @memberof repository
         * @interface ICollectionDependency
         * @property {string|null} [id] CollectionDependency id
         * @property {number|Long|null} [mtime] CollectionDependency mtime
         * @property {string|null} [asset_id] CollectionDependency asset_id
         * @property {string|null} [dependency_id] CollectionDependency dependency_id
         * @property {string|null} [dependency_type_id] CollectionDependency dependency_type_id
         * @property {boolean|null} [synced] CollectionDependency synced
         */

        /**
         * Constructs a new CollectionDependency.
         * @memberof repository
         * @classdesc Represents a CollectionDependency.
         * @implements ICollectionDependency
         * @constructor
         * @param {repository.ICollectionDependency=} [properties] Properties to set
         */
        function CollectionDependency(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * CollectionDependency id.
         * @member {string} id
         * @memberof repository.CollectionDependency
         * @instance
         */
        CollectionDependency.prototype.id = "";

        /**
         * CollectionDependency mtime.
         * @member {number|Long} mtime
         * @memberof repository.CollectionDependency
         * @instance
         */
        CollectionDependency.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * CollectionDependency asset_id.
         * @member {string} asset_id
         * @memberof repository.CollectionDependency
         * @instance
         */
        CollectionDependency.prototype.asset_id = "";

        /**
         * CollectionDependency dependency_id.
         * @member {string} dependency_id
         * @memberof repository.CollectionDependency
         * @instance
         */
        CollectionDependency.prototype.dependency_id = "";

        /**
         * CollectionDependency dependency_type_id.
         * @member {string} dependency_type_id
         * @memberof repository.CollectionDependency
         * @instance
         */
        CollectionDependency.prototype.dependency_type_id = "";

        /**
         * CollectionDependency synced.
         * @member {boolean} synced
         * @memberof repository.CollectionDependency
         * @instance
         */
        CollectionDependency.prototype.synced = false;

        /**
         * Creates a new CollectionDependency instance using the specified properties.
         * @function create
         * @memberof repository.CollectionDependency
         * @static
         * @param {repository.ICollectionDependency=} [properties] Properties to set
         * @returns {repository.CollectionDependency} CollectionDependency instance
         */
        CollectionDependency.create = function create(properties) {
            return new CollectionDependency(properties);
        };

        /**
         * Encodes the specified CollectionDependency message. Does not implicitly {@link repository.CollectionDependency.verify|verify} messages.
         * @function encode
         * @memberof repository.CollectionDependency
         * @static
         * @param {repository.ICollectionDependency} message CollectionDependency message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CollectionDependency.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.asset_id != null && Object.hasOwnProperty.call(message, "asset_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.asset_id);
            if (message.dependency_id != null && Object.hasOwnProperty.call(message, "dependency_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.dependency_id);
            if (message.dependency_type_id != null && Object.hasOwnProperty.call(message, "dependency_type_id"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.dependency_type_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified CollectionDependency message, length delimited. Does not implicitly {@link repository.CollectionDependency.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.CollectionDependency
         * @static
         * @param {repository.ICollectionDependency} message CollectionDependency message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        CollectionDependency.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a CollectionDependency message from the specified reader or buffer.
         * @function decode
         * @memberof repository.CollectionDependency
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.CollectionDependency} CollectionDependency
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CollectionDependency.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.CollectionDependency();
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
                        message.asset_id = reader.string();
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
         * Decodes a CollectionDependency message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.CollectionDependency
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.CollectionDependency} CollectionDependency
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        CollectionDependency.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a CollectionDependency message.
         * @function verify
         * @memberof repository.CollectionDependency
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        CollectionDependency.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.asset_id != null && message.hasOwnProperty("asset_id"))
                if (!$util.isString(message.asset_id))
                    return "asset_id: string expected";
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
         * Creates a CollectionDependency message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.CollectionDependency
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.CollectionDependency} CollectionDependency
         */
        CollectionDependency.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.CollectionDependency)
                return object;
            let message = new $root.repository.CollectionDependency();
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
            if (object.asset_id != null)
                message.asset_id = String(object.asset_id);
            if (object.dependency_id != null)
                message.dependency_id = String(object.dependency_id);
            if (object.dependency_type_id != null)
                message.dependency_type_id = String(object.dependency_type_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a CollectionDependency message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.CollectionDependency
         * @static
         * @param {repository.CollectionDependency} message CollectionDependency
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        CollectionDependency.toObject = function toObject(message, options) {
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
                object.asset_id = "";
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
            if (message.asset_id != null && message.hasOwnProperty("asset_id"))
                object.asset_id = message.asset_id;
            if (message.dependency_id != null && message.hasOwnProperty("dependency_id"))
                object.dependency_id = message.dependency_id;
            if (message.dependency_type_id != null && message.hasOwnProperty("dependency_type_id"))
                object.dependency_type_id = message.dependency_type_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this CollectionDependency to JSON.
         * @function toJSON
         * @memberof repository.CollectionDependency
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        CollectionDependency.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for CollectionDependency
         * @function getTypeUrl
         * @memberof repository.CollectionDependency
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        CollectionDependency.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.CollectionDependency";
        };

        return CollectionDependency;
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

    repository.WorkflowAsset = (function() {

        /**
         * Properties of a WorkflowAsset.
         * @memberof repository
         * @interface IWorkflowAsset
         * @property {string|null} [id] WorkflowAsset id
         * @property {number|Long|null} [mtime] WorkflowAsset mtime
         * @property {string|null} [name] WorkflowAsset name
         * @property {string|null} [template_id] WorkflowAsset template_id
         * @property {boolean|null} [is_resource] WorkflowAsset is_resource
         * @property {string|null} [workflow_id] WorkflowAsset workflow_id
         * @property {string|null} [asset_type_id] WorkflowAsset asset_type_id
         * @property {string|null} [workflow_collection_id] WorkflowAsset workflow_collection_id
         * @property {boolean|null} [is_link] WorkflowAsset is_link
         * @property {string|null} [pointer] WorkflowAsset pointer
         * @property {boolean|null} [synced] WorkflowAsset synced
         */

        /**
         * Constructs a new WorkflowAsset.
         * @memberof repository
         * @classdesc Represents a WorkflowAsset.
         * @implements IWorkflowAsset
         * @constructor
         * @param {repository.IWorkflowAsset=} [properties] Properties to set
         */
        function WorkflowAsset(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WorkflowAsset id.
         * @member {string} id
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.id = "";

        /**
         * WorkflowAsset mtime.
         * @member {number|Long} mtime
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * WorkflowAsset name.
         * @member {string} name
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.name = "";

        /**
         * WorkflowAsset template_id.
         * @member {string} template_id
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.template_id = "";

        /**
         * WorkflowAsset is_resource.
         * @member {boolean} is_resource
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.is_resource = false;

        /**
         * WorkflowAsset workflow_id.
         * @member {string} workflow_id
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.workflow_id = "";

        /**
         * WorkflowAsset asset_type_id.
         * @member {string} asset_type_id
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.asset_type_id = "";

        /**
         * WorkflowAsset workflow_collection_id.
         * @member {string} workflow_collection_id
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.workflow_collection_id = "";

        /**
         * WorkflowAsset is_link.
         * @member {boolean} is_link
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.is_link = false;

        /**
         * WorkflowAsset pointer.
         * @member {string} pointer
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.pointer = "";

        /**
         * WorkflowAsset synced.
         * @member {boolean} synced
         * @memberof repository.WorkflowAsset
         * @instance
         */
        WorkflowAsset.prototype.synced = false;

        /**
         * Creates a new WorkflowAsset instance using the specified properties.
         * @function create
         * @memberof repository.WorkflowAsset
         * @static
         * @param {repository.IWorkflowAsset=} [properties] Properties to set
         * @returns {repository.WorkflowAsset} WorkflowAsset instance
         */
        WorkflowAsset.create = function create(properties) {
            return new WorkflowAsset(properties);
        };

        /**
         * Encodes the specified WorkflowAsset message. Does not implicitly {@link repository.WorkflowAsset.verify|verify} messages.
         * @function encode
         * @memberof repository.WorkflowAsset
         * @static
         * @param {repository.IWorkflowAsset} message WorkflowAsset message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WorkflowAsset.encode = function encode(message, writer) {
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
            if (message.asset_type_id != null && Object.hasOwnProperty.call(message, "asset_type_id"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.asset_type_id);
            if (message.workflow_collection_id != null && Object.hasOwnProperty.call(message, "workflow_collection_id"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.workflow_collection_id);
            if (message.is_link != null && Object.hasOwnProperty.call(message, "is_link"))
                writer.uint32(/* id 9, wireType 0 =*/72).bool(message.is_link);
            if (message.pointer != null && Object.hasOwnProperty.call(message, "pointer"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.pointer);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 11, wireType 0 =*/88).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified WorkflowAsset message, length delimited. Does not implicitly {@link repository.WorkflowAsset.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.WorkflowAsset
         * @static
         * @param {repository.IWorkflowAsset} message WorkflowAsset message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WorkflowAsset.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a WorkflowAsset message from the specified reader or buffer.
         * @function decode
         * @memberof repository.WorkflowAsset
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.WorkflowAsset} WorkflowAsset
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WorkflowAsset.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.WorkflowAsset();
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
                        message.asset_type_id = reader.string();
                        break;
                    }
                case 8: {
                        message.workflow_collection_id = reader.string();
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
         * Decodes a WorkflowAsset message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.WorkflowAsset
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.WorkflowAsset} WorkflowAsset
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WorkflowAsset.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a WorkflowAsset message.
         * @function verify
         * @memberof repository.WorkflowAsset
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        WorkflowAsset.verify = function verify(message) {
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
            if (message.asset_type_id != null && message.hasOwnProperty("asset_type_id"))
                if (!$util.isString(message.asset_type_id))
                    return "asset_type_id: string expected";
            if (message.workflow_collection_id != null && message.hasOwnProperty("workflow_collection_id"))
                if (!$util.isString(message.workflow_collection_id))
                    return "workflow_collection_id: string expected";
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
         * Creates a WorkflowAsset message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.WorkflowAsset
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.WorkflowAsset} WorkflowAsset
         */
        WorkflowAsset.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.WorkflowAsset)
                return object;
            let message = new $root.repository.WorkflowAsset();
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
            if (object.asset_type_id != null)
                message.asset_type_id = String(object.asset_type_id);
            if (object.workflow_collection_id != null)
                message.workflow_collection_id = String(object.workflow_collection_id);
            if (object.is_link != null)
                message.is_link = Boolean(object.is_link);
            if (object.pointer != null)
                message.pointer = String(object.pointer);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a WorkflowAsset message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.WorkflowAsset
         * @static
         * @param {repository.WorkflowAsset} message WorkflowAsset
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        WorkflowAsset.toObject = function toObject(message, options) {
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
                object.asset_type_id = "";
                object.workflow_collection_id = "";
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
            if (message.asset_type_id != null && message.hasOwnProperty("asset_type_id"))
                object.asset_type_id = message.asset_type_id;
            if (message.workflow_collection_id != null && message.hasOwnProperty("workflow_collection_id"))
                object.workflow_collection_id = message.workflow_collection_id;
            if (message.is_link != null && message.hasOwnProperty("is_link"))
                object.is_link = message.is_link;
            if (message.pointer != null && message.hasOwnProperty("pointer"))
                object.pointer = message.pointer;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this WorkflowAsset to JSON.
         * @function toJSON
         * @memberof repository.WorkflowAsset
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        WorkflowAsset.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for WorkflowAsset
         * @function getTypeUrl
         * @memberof repository.WorkflowAsset
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WorkflowAsset.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.WorkflowAsset";
        };

        return WorkflowAsset;
    })();

    repository.WorkflowCollection = (function() {

        /**
         * Properties of a WorkflowCollection.
         * @memberof repository
         * @interface IWorkflowCollection
         * @property {string|null} [id] WorkflowCollection id
         * @property {number|Long|null} [mtime] WorkflowCollection mtime
         * @property {string|null} [name] WorkflowCollection name
         * @property {string|null} [workflow_id] WorkflowCollection workflow_id
         * @property {string|null} [collection_type_id] WorkflowCollection collection_type_id
         * @property {string|null} [parent_id] WorkflowCollection parent_id
         * @property {boolean|null} [synced] WorkflowCollection synced
         */

        /**
         * Constructs a new WorkflowCollection.
         * @memberof repository
         * @classdesc Represents a WorkflowCollection.
         * @implements IWorkflowCollection
         * @constructor
         * @param {repository.IWorkflowCollection=} [properties] Properties to set
         */
        function WorkflowCollection(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * WorkflowCollection id.
         * @member {string} id
         * @memberof repository.WorkflowCollection
         * @instance
         */
        WorkflowCollection.prototype.id = "";

        /**
         * WorkflowCollection mtime.
         * @member {number|Long} mtime
         * @memberof repository.WorkflowCollection
         * @instance
         */
        WorkflowCollection.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * WorkflowCollection name.
         * @member {string} name
         * @memberof repository.WorkflowCollection
         * @instance
         */
        WorkflowCollection.prototype.name = "";

        /**
         * WorkflowCollection workflow_id.
         * @member {string} workflow_id
         * @memberof repository.WorkflowCollection
         * @instance
         */
        WorkflowCollection.prototype.workflow_id = "";

        /**
         * WorkflowCollection collection_type_id.
         * @member {string} collection_type_id
         * @memberof repository.WorkflowCollection
         * @instance
         */
        WorkflowCollection.prototype.collection_type_id = "";

        /**
         * WorkflowCollection parent_id.
         * @member {string} parent_id
         * @memberof repository.WorkflowCollection
         * @instance
         */
        WorkflowCollection.prototype.parent_id = "";

        /**
         * WorkflowCollection synced.
         * @member {boolean} synced
         * @memberof repository.WorkflowCollection
         * @instance
         */
        WorkflowCollection.prototype.synced = false;

        /**
         * Creates a new WorkflowCollection instance using the specified properties.
         * @function create
         * @memberof repository.WorkflowCollection
         * @static
         * @param {repository.IWorkflowCollection=} [properties] Properties to set
         * @returns {repository.WorkflowCollection} WorkflowCollection instance
         */
        WorkflowCollection.create = function create(properties) {
            return new WorkflowCollection(properties);
        };

        /**
         * Encodes the specified WorkflowCollection message. Does not implicitly {@link repository.WorkflowCollection.verify|verify} messages.
         * @function encode
         * @memberof repository.WorkflowCollection
         * @static
         * @param {repository.IWorkflowCollection} message WorkflowCollection message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WorkflowCollection.encode = function encode(message, writer) {
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
            if (message.collection_type_id != null && Object.hasOwnProperty.call(message, "collection_type_id"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.collection_type_id);
            if (message.parent_id != null && Object.hasOwnProperty.call(message, "parent_id"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.parent_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 7, wireType 0 =*/56).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified WorkflowCollection message, length delimited. Does not implicitly {@link repository.WorkflowCollection.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.WorkflowCollection
         * @static
         * @param {repository.IWorkflowCollection} message WorkflowCollection message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        WorkflowCollection.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a WorkflowCollection message from the specified reader or buffer.
         * @function decode
         * @memberof repository.WorkflowCollection
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.WorkflowCollection} WorkflowCollection
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WorkflowCollection.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.WorkflowCollection();
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
                        message.collection_type_id = reader.string();
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
         * Decodes a WorkflowCollection message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.WorkflowCollection
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.WorkflowCollection} WorkflowCollection
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        WorkflowCollection.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a WorkflowCollection message.
         * @function verify
         * @memberof repository.WorkflowCollection
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        WorkflowCollection.verify = function verify(message) {
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
            if (message.collection_type_id != null && message.hasOwnProperty("collection_type_id"))
                if (!$util.isString(message.collection_type_id))
                    return "collection_type_id: string expected";
            if (message.parent_id != null && message.hasOwnProperty("parent_id"))
                if (!$util.isString(message.parent_id))
                    return "parent_id: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates a WorkflowCollection message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.WorkflowCollection
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.WorkflowCollection} WorkflowCollection
         */
        WorkflowCollection.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.WorkflowCollection)
                return object;
            let message = new $root.repository.WorkflowCollection();
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
            if (object.collection_type_id != null)
                message.collection_type_id = String(object.collection_type_id);
            if (object.parent_id != null)
                message.parent_id = String(object.parent_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from a WorkflowCollection message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.WorkflowCollection
         * @static
         * @param {repository.WorkflowCollection} message WorkflowCollection
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        WorkflowCollection.toObject = function toObject(message, options) {
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
                object.collection_type_id = "";
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
            if (message.collection_type_id != null && message.hasOwnProperty("collection_type_id"))
                object.collection_type_id = message.collection_type_id;
            if (message.parent_id != null && message.hasOwnProperty("parent_id"))
                object.parent_id = message.parent_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this WorkflowCollection to JSON.
         * @function toJSON
         * @memberof repository.WorkflowCollection
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        WorkflowCollection.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for WorkflowCollection
         * @function getTypeUrl
         * @memberof repository.WorkflowCollection
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        WorkflowCollection.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.WorkflowCollection";
        };

        return WorkflowCollection;
    })();

    repository.WorkflowLink = (function() {

        /**
         * Properties of a WorkflowLink.
         * @memberof repository
         * @interface IWorkflowLink
         * @property {string|null} [id] WorkflowLink id
         * @property {number|Long|null} [mtime] WorkflowLink mtime
         * @property {string|null} [name] WorkflowLink name
         * @property {string|null} [collection_type_id] WorkflowLink collection_type_id
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
         * WorkflowLink collection_type_id.
         * @member {string} collection_type_id
         * @memberof repository.WorkflowLink
         * @instance
         */
        WorkflowLink.prototype.collection_type_id = "";

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
            if (message.collection_type_id != null && Object.hasOwnProperty.call(message, "collection_type_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.collection_type_id);
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
                        message.collection_type_id = reader.string();
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
            if (message.collection_type_id != null && message.hasOwnProperty("collection_type_id"))
                if (!$util.isString(message.collection_type_id))
                    return "collection_type_id: string expected";
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
            if (object.collection_type_id != null)
                message.collection_type_id = String(object.collection_type_id);
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
                object.collection_type_id = "";
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
            if (message.collection_type_id != null && message.hasOwnProperty("collection_type_id"))
                object.collection_type_id = message.collection_type_id;
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

    repository.AssetTag = (function() {

        /**
         * Properties of an AssetTag.
         * @memberof repository
         * @interface IAssetTag
         * @property {string|null} [id] AssetTag id
         * @property {number|Long|null} [mtime] AssetTag mtime
         * @property {string|null} [asset_id] AssetTag asset_id
         * @property {string|null} [tag_id] AssetTag tag_id
         * @property {boolean|null} [synced] AssetTag synced
         */

        /**
         * Constructs a new AssetTag.
         * @memberof repository
         * @classdesc Represents an AssetTag.
         * @implements IAssetTag
         * @constructor
         * @param {repository.IAssetTag=} [properties] Properties to set
         */
        function AssetTag(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * AssetTag id.
         * @member {string} id
         * @memberof repository.AssetTag
         * @instance
         */
        AssetTag.prototype.id = "";

        /**
         * AssetTag mtime.
         * @member {number|Long} mtime
         * @memberof repository.AssetTag
         * @instance
         */
        AssetTag.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * AssetTag asset_id.
         * @member {string} asset_id
         * @memberof repository.AssetTag
         * @instance
         */
        AssetTag.prototype.asset_id = "";

        /**
         * AssetTag tag_id.
         * @member {string} tag_id
         * @memberof repository.AssetTag
         * @instance
         */
        AssetTag.prototype.tag_id = "";

        /**
         * AssetTag synced.
         * @member {boolean} synced
         * @memberof repository.AssetTag
         * @instance
         */
        AssetTag.prototype.synced = false;

        /**
         * Creates a new AssetTag instance using the specified properties.
         * @function create
         * @memberof repository.AssetTag
         * @static
         * @param {repository.IAssetTag=} [properties] Properties to set
         * @returns {repository.AssetTag} AssetTag instance
         */
        AssetTag.create = function create(properties) {
            return new AssetTag(properties);
        };

        /**
         * Encodes the specified AssetTag message. Does not implicitly {@link repository.AssetTag.verify|verify} messages.
         * @function encode
         * @memberof repository.AssetTag
         * @static
         * @param {repository.IAssetTag} message AssetTag message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AssetTag.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.asset_id != null && Object.hasOwnProperty.call(message, "asset_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.asset_id);
            if (message.tag_id != null && Object.hasOwnProperty.call(message, "tag_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.tag_id);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 5, wireType 0 =*/40).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified AssetTag message, length delimited. Does not implicitly {@link repository.AssetTag.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.AssetTag
         * @static
         * @param {repository.IAssetTag} message AssetTag message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        AssetTag.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an AssetTag message from the specified reader or buffer.
         * @function decode
         * @memberof repository.AssetTag
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.AssetTag} AssetTag
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AssetTag.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.AssetTag();
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
                        message.asset_id = reader.string();
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
         * Decodes an AssetTag message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.AssetTag
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.AssetTag} AssetTag
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        AssetTag.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an AssetTag message.
         * @function verify
         * @memberof repository.AssetTag
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        AssetTag.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.asset_id != null && message.hasOwnProperty("asset_id"))
                if (!$util.isString(message.asset_id))
                    return "asset_id: string expected";
            if (message.tag_id != null && message.hasOwnProperty("tag_id"))
                if (!$util.isString(message.tag_id))
                    return "tag_id: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates an AssetTag message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.AssetTag
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.AssetTag} AssetTag
         */
        AssetTag.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.AssetTag)
                return object;
            let message = new $root.repository.AssetTag();
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
            if (object.asset_id != null)
                message.asset_id = String(object.asset_id);
            if (object.tag_id != null)
                message.tag_id = String(object.tag_id);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from an AssetTag message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.AssetTag
         * @static
         * @param {repository.AssetTag} message AssetTag
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        AssetTag.toObject = function toObject(message, options) {
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
                object.asset_id = "";
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
            if (message.asset_id != null && message.hasOwnProperty("asset_id"))
                object.asset_id = message.asset_id;
            if (message.tag_id != null && message.hasOwnProperty("tag_id"))
                object.tag_id = message.tag_id;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this AssetTag to JSON.
         * @function toJSON
         * @memberof repository.AssetTag
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        AssetTag.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for AssetTag
         * @function getTypeUrl
         * @memberof repository.AssetTag
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        AssetTag.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.AssetTag";
        };

        return AssetTag;
    })();

    repository.Checkpoint = (function() {

        /**
         * Properties of a Checkpoint.
         * @memberof repository
         * @interface ICheckpoint
         * @property {string|null} [id] Checkpoint id
         * @property {number|Long|null} [mtime] Checkpoint mtime
         * @property {string|null} [created_at] Checkpoint created_at
         * @property {string|null} [asset_id] Checkpoint asset_id
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
         * Checkpoint asset_id.
         * @member {string} asset_id
         * @memberof repository.Checkpoint
         * @instance
         */
        Checkpoint.prototype.asset_id = "";

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
            if (message.asset_id != null && Object.hasOwnProperty.call(message, "asset_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.asset_id);
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
                        message.asset_id = reader.string();
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
            if (message.asset_id != null && message.hasOwnProperty("asset_id"))
                if (!$util.isString(message.asset_id))
                    return "asset_id: string expected";
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
            if (object.asset_id != null)
                message.asset_id = String(object.asset_id);
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
                object.asset_id = "";
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
            if (message.asset_id != null && message.hasOwnProperty("asset_id"))
                object.asset_id = message.asset_id;
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
         * @property {boolean|null} [view_collection] Role view_collection
         * @property {boolean|null} [create_collection] Role create_collection
         * @property {boolean|null} [update_collection] Role update_collection
         * @property {boolean|null} [delete_collection] Role delete_collection
         * @property {boolean|null} [view_asset] Role view_asset
         * @property {boolean|null} [create_asset] Role create_asset
         * @property {boolean|null} [update_asset] Role update_asset
         * @property {boolean|null} [delete_asset] Role delete_asset
         * @property {boolean|null} [view_template] Role view_template
         * @property {boolean|null} [create_template] Role create_template
         * @property {boolean|null} [update_template] Role update_template
         * @property {boolean|null} [delete_template] Role delete_template
         * @property {boolean|null} [view_checkpoint] Role view_checkpoint
         * @property {boolean|null} [create_checkpoint] Role create_checkpoint
         * @property {boolean|null} [delete_checkpoint] Role delete_checkpoint
         * @property {boolean|null} [pull_chunk] Role pull_chunk
         * @property {boolean|null} [assign_asset] Role assign_asset
         * @property {boolean|null} [unassign_asset] Role unassign_asset
         * @property {boolean|null} [add_user] Role add_user
         * @property {boolean|null} [remove_user] Role remove_user
         * @property {boolean|null} [change_role] Role change_role
         * @property {boolean|null} [change_status] Role change_status
         * @property {boolean|null} [set_done_asset] Role set_done_asset
         * @property {boolean|null} [set_retake_asset] Role set_retake_asset
         * @property {boolean|null} [view_done_asset] Role view_done_asset
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
         * Role view_collection.
         * @member {boolean} view_collection
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.view_collection = false;

        /**
         * Role create_collection.
         * @member {boolean} create_collection
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.create_collection = false;

        /**
         * Role update_collection.
         * @member {boolean} update_collection
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.update_collection = false;

        /**
         * Role delete_collection.
         * @member {boolean} delete_collection
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.delete_collection = false;

        /**
         * Role view_asset.
         * @member {boolean} view_asset
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.view_asset = false;

        /**
         * Role create_asset.
         * @member {boolean} create_asset
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.create_asset = false;

        /**
         * Role update_asset.
         * @member {boolean} update_asset
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.update_asset = false;

        /**
         * Role delete_asset.
         * @member {boolean} delete_asset
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.delete_asset = false;

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
         * Role assign_asset.
         * @member {boolean} assign_asset
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.assign_asset = false;

        /**
         * Role unassign_asset.
         * @member {boolean} unassign_asset
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.unassign_asset = false;

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
         * Role set_done_asset.
         * @member {boolean} set_done_asset
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.set_done_asset = false;

        /**
         * Role set_retake_asset.
         * @member {boolean} set_retake_asset
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.set_retake_asset = false;

        /**
         * Role view_done_asset.
         * @member {boolean} view_done_asset
         * @memberof repository.Role
         * @instance
         */
        Role.prototype.view_done_asset = false;

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
            if (message.view_collection != null && Object.hasOwnProperty.call(message, "view_collection"))
                writer.uint32(/* id 5, wireType 0 =*/40).bool(message.view_collection);
            if (message.create_collection != null && Object.hasOwnProperty.call(message, "create_collection"))
                writer.uint32(/* id 6, wireType 0 =*/48).bool(message.create_collection);
            if (message.update_collection != null && Object.hasOwnProperty.call(message, "update_collection"))
                writer.uint32(/* id 7, wireType 0 =*/56).bool(message.update_collection);
            if (message.delete_collection != null && Object.hasOwnProperty.call(message, "delete_collection"))
                writer.uint32(/* id 8, wireType 0 =*/64).bool(message.delete_collection);
            if (message.view_asset != null && Object.hasOwnProperty.call(message, "view_asset"))
                writer.uint32(/* id 9, wireType 0 =*/72).bool(message.view_asset);
            if (message.create_asset != null && Object.hasOwnProperty.call(message, "create_asset"))
                writer.uint32(/* id 10, wireType 0 =*/80).bool(message.create_asset);
            if (message.update_asset != null && Object.hasOwnProperty.call(message, "update_asset"))
                writer.uint32(/* id 11, wireType 0 =*/88).bool(message.update_asset);
            if (message.delete_asset != null && Object.hasOwnProperty.call(message, "delete_asset"))
                writer.uint32(/* id 12, wireType 0 =*/96).bool(message.delete_asset);
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
            if (message.assign_asset != null && Object.hasOwnProperty.call(message, "assign_asset"))
                writer.uint32(/* id 21, wireType 0 =*/168).bool(message.assign_asset);
            if (message.unassign_asset != null && Object.hasOwnProperty.call(message, "unassign_asset"))
                writer.uint32(/* id 22, wireType 0 =*/176).bool(message.unassign_asset);
            if (message.add_user != null && Object.hasOwnProperty.call(message, "add_user"))
                writer.uint32(/* id 23, wireType 0 =*/184).bool(message.add_user);
            if (message.remove_user != null && Object.hasOwnProperty.call(message, "remove_user"))
                writer.uint32(/* id 24, wireType 0 =*/192).bool(message.remove_user);
            if (message.change_role != null && Object.hasOwnProperty.call(message, "change_role"))
                writer.uint32(/* id 25, wireType 0 =*/200).bool(message.change_role);
            if (message.change_status != null && Object.hasOwnProperty.call(message, "change_status"))
                writer.uint32(/* id 26, wireType 0 =*/208).bool(message.change_status);
            if (message.set_done_asset != null && Object.hasOwnProperty.call(message, "set_done_asset"))
                writer.uint32(/* id 27, wireType 0 =*/216).bool(message.set_done_asset);
            if (message.set_retake_asset != null && Object.hasOwnProperty.call(message, "set_retake_asset"))
                writer.uint32(/* id 28, wireType 0 =*/224).bool(message.set_retake_asset);
            if (message.view_done_asset != null && Object.hasOwnProperty.call(message, "view_done_asset"))
                writer.uint32(/* id 29, wireType 0 =*/232).bool(message.view_done_asset);
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
                        message.view_collection = reader.bool();
                        break;
                    }
                case 6: {
                        message.create_collection = reader.bool();
                        break;
                    }
                case 7: {
                        message.update_collection = reader.bool();
                        break;
                    }
                case 8: {
                        message.delete_collection = reader.bool();
                        break;
                    }
                case 9: {
                        message.view_asset = reader.bool();
                        break;
                    }
                case 10: {
                        message.create_asset = reader.bool();
                        break;
                    }
                case 11: {
                        message.update_asset = reader.bool();
                        break;
                    }
                case 12: {
                        message.delete_asset = reader.bool();
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
                        message.assign_asset = reader.bool();
                        break;
                    }
                case 22: {
                        message.unassign_asset = reader.bool();
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
                        message.set_done_asset = reader.bool();
                        break;
                    }
                case 28: {
                        message.set_retake_asset = reader.bool();
                        break;
                    }
                case 29: {
                        message.view_done_asset = reader.bool();
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
            if (message.view_collection != null && message.hasOwnProperty("view_collection"))
                if (typeof message.view_collection !== "boolean")
                    return "view_collection: boolean expected";
            if (message.create_collection != null && message.hasOwnProperty("create_collection"))
                if (typeof message.create_collection !== "boolean")
                    return "create_collection: boolean expected";
            if (message.update_collection != null && message.hasOwnProperty("update_collection"))
                if (typeof message.update_collection !== "boolean")
                    return "update_collection: boolean expected";
            if (message.delete_collection != null && message.hasOwnProperty("delete_collection"))
                if (typeof message.delete_collection !== "boolean")
                    return "delete_collection: boolean expected";
            if (message.view_asset != null && message.hasOwnProperty("view_asset"))
                if (typeof message.view_asset !== "boolean")
                    return "view_asset: boolean expected";
            if (message.create_asset != null && message.hasOwnProperty("create_asset"))
                if (typeof message.create_asset !== "boolean")
                    return "create_asset: boolean expected";
            if (message.update_asset != null && message.hasOwnProperty("update_asset"))
                if (typeof message.update_asset !== "boolean")
                    return "update_asset: boolean expected";
            if (message.delete_asset != null && message.hasOwnProperty("delete_asset"))
                if (typeof message.delete_asset !== "boolean")
                    return "delete_asset: boolean expected";
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
            if (message.assign_asset != null && message.hasOwnProperty("assign_asset"))
                if (typeof message.assign_asset !== "boolean")
                    return "assign_asset: boolean expected";
            if (message.unassign_asset != null && message.hasOwnProperty("unassign_asset"))
                if (typeof message.unassign_asset !== "boolean")
                    return "unassign_asset: boolean expected";
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
            if (message.set_done_asset != null && message.hasOwnProperty("set_done_asset"))
                if (typeof message.set_done_asset !== "boolean")
                    return "set_done_asset: boolean expected";
            if (message.set_retake_asset != null && message.hasOwnProperty("set_retake_asset"))
                if (typeof message.set_retake_asset !== "boolean")
                    return "set_retake_asset: boolean expected";
            if (message.view_done_asset != null && message.hasOwnProperty("view_done_asset"))
                if (typeof message.view_done_asset !== "boolean")
                    return "view_done_asset: boolean expected";
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
            if (object.view_collection != null)
                message.view_collection = Boolean(object.view_collection);
            if (object.create_collection != null)
                message.create_collection = Boolean(object.create_collection);
            if (object.update_collection != null)
                message.update_collection = Boolean(object.update_collection);
            if (object.delete_collection != null)
                message.delete_collection = Boolean(object.delete_collection);
            if (object.view_asset != null)
                message.view_asset = Boolean(object.view_asset);
            if (object.create_asset != null)
                message.create_asset = Boolean(object.create_asset);
            if (object.update_asset != null)
                message.update_asset = Boolean(object.update_asset);
            if (object.delete_asset != null)
                message.delete_asset = Boolean(object.delete_asset);
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
            if (object.assign_asset != null)
                message.assign_asset = Boolean(object.assign_asset);
            if (object.unassign_asset != null)
                message.unassign_asset = Boolean(object.unassign_asset);
            if (object.add_user != null)
                message.add_user = Boolean(object.add_user);
            if (object.remove_user != null)
                message.remove_user = Boolean(object.remove_user);
            if (object.change_role != null)
                message.change_role = Boolean(object.change_role);
            if (object.change_status != null)
                message.change_status = Boolean(object.change_status);
            if (object.set_done_asset != null)
                message.set_done_asset = Boolean(object.set_done_asset);
            if (object.set_retake_asset != null)
                message.set_retake_asset = Boolean(object.set_retake_asset);
            if (object.view_done_asset != null)
                message.view_done_asset = Boolean(object.view_done_asset);
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
                object.view_collection = false;
                object.create_collection = false;
                object.update_collection = false;
                object.delete_collection = false;
                object.view_asset = false;
                object.create_asset = false;
                object.update_asset = false;
                object.delete_asset = false;
                object.view_template = false;
                object.create_template = false;
                object.update_template = false;
                object.delete_template = false;
                object.view_checkpoint = false;
                object.create_checkpoint = false;
                object.delete_checkpoint = false;
                object.pull_chunk = false;
                object.assign_asset = false;
                object.unassign_asset = false;
                object.add_user = false;
                object.remove_user = false;
                object.change_role = false;
                object.change_status = false;
                object.set_done_asset = false;
                object.set_retake_asset = false;
                object.view_done_asset = false;
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
            if (message.view_collection != null && message.hasOwnProperty("view_collection"))
                object.view_collection = message.view_collection;
            if (message.create_collection != null && message.hasOwnProperty("create_collection"))
                object.create_collection = message.create_collection;
            if (message.update_collection != null && message.hasOwnProperty("update_collection"))
                object.update_collection = message.update_collection;
            if (message.delete_collection != null && message.hasOwnProperty("delete_collection"))
                object.delete_collection = message.delete_collection;
            if (message.view_asset != null && message.hasOwnProperty("view_asset"))
                object.view_asset = message.view_asset;
            if (message.create_asset != null && message.hasOwnProperty("create_asset"))
                object.create_asset = message.create_asset;
            if (message.update_asset != null && message.hasOwnProperty("update_asset"))
                object.update_asset = message.update_asset;
            if (message.delete_asset != null && message.hasOwnProperty("delete_asset"))
                object.delete_asset = message.delete_asset;
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
            if (message.assign_asset != null && message.hasOwnProperty("assign_asset"))
                object.assign_asset = message.assign_asset;
            if (message.unassign_asset != null && message.hasOwnProperty("unassign_asset"))
                object.unassign_asset = message.unassign_asset;
            if (message.add_user != null && message.hasOwnProperty("add_user"))
                object.add_user = message.add_user;
            if (message.remove_user != null && message.hasOwnProperty("remove_user"))
                object.remove_user = message.remove_user;
            if (message.change_role != null && message.hasOwnProperty("change_role"))
                object.change_role = message.change_role;
            if (message.change_status != null && message.hasOwnProperty("change_status"))
                object.change_status = message.change_status;
            if (message.set_done_asset != null && message.hasOwnProperty("set_done_asset"))
                object.set_done_asset = message.set_done_asset;
            if (message.set_retake_asset != null && message.hasOwnProperty("set_retake_asset"))
                object.set_retake_asset = message.set_retake_asset;
            if (message.view_done_asset != null && message.hasOwnProperty("view_done_asset"))
                object.view_done_asset = message.view_done_asset;
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

    repository.IntegrationProject = (function() {

        /**
         * Properties of an IntegrationProject.
         * @memberof repository
         * @interface IIntegrationProject
         * @property {string|null} [id] IntegrationProject id
         * @property {number|Long|null} [mtime] IntegrationProject mtime
         * @property {string|null} [integration_id] IntegrationProject integration_id
         * @property {string|null} [external_project_id] IntegrationProject external_project_id
         * @property {string|null} [external_project_name] IntegrationProject external_project_name
         * @property {string|null} [api_url] IntegrationProject api_url
         * @property {string|null} [sync_options] IntegrationProject sync_options
         * @property {string|null} [linked_by_user_id] IntegrationProject linked_by_user_id
         * @property {string|null} [linked_at] IntegrationProject linked_at
         * @property {boolean|null} [enabled] IntegrationProject enabled
         * @property {boolean|null} [synced] IntegrationProject synced
         */

        /**
         * Constructs a new IntegrationProject.
         * @memberof repository
         * @classdesc Represents an IntegrationProject.
         * @implements IIntegrationProject
         * @constructor
         * @param {repository.IIntegrationProject=} [properties] Properties to set
         */
        function IntegrationProject(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * IntegrationProject id.
         * @member {string} id
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.id = "";

        /**
         * IntegrationProject mtime.
         * @member {number|Long} mtime
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * IntegrationProject integration_id.
         * @member {string} integration_id
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.integration_id = "";

        /**
         * IntegrationProject external_project_id.
         * @member {string} external_project_id
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.external_project_id = "";

        /**
         * IntegrationProject external_project_name.
         * @member {string} external_project_name
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.external_project_name = "";

        /**
         * IntegrationProject api_url.
         * @member {string} api_url
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.api_url = "";

        /**
         * IntegrationProject sync_options.
         * @member {string} sync_options
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.sync_options = "";

        /**
         * IntegrationProject linked_by_user_id.
         * @member {string} linked_by_user_id
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.linked_by_user_id = "";

        /**
         * IntegrationProject linked_at.
         * @member {string} linked_at
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.linked_at = "";

        /**
         * IntegrationProject enabled.
         * @member {boolean} enabled
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.enabled = false;

        /**
         * IntegrationProject synced.
         * @member {boolean} synced
         * @memberof repository.IntegrationProject
         * @instance
         */
        IntegrationProject.prototype.synced = false;

        /**
         * Creates a new IntegrationProject instance using the specified properties.
         * @function create
         * @memberof repository.IntegrationProject
         * @static
         * @param {repository.IIntegrationProject=} [properties] Properties to set
         * @returns {repository.IntegrationProject} IntegrationProject instance
         */
        IntegrationProject.create = function create(properties) {
            return new IntegrationProject(properties);
        };

        /**
         * Encodes the specified IntegrationProject message. Does not implicitly {@link repository.IntegrationProject.verify|verify} messages.
         * @function encode
         * @memberof repository.IntegrationProject
         * @static
         * @param {repository.IIntegrationProject} message IntegrationProject message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        IntegrationProject.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.integration_id != null && Object.hasOwnProperty.call(message, "integration_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.integration_id);
            if (message.external_project_id != null && Object.hasOwnProperty.call(message, "external_project_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.external_project_id);
            if (message.external_project_name != null && Object.hasOwnProperty.call(message, "external_project_name"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.external_project_name);
            if (message.api_url != null && Object.hasOwnProperty.call(message, "api_url"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.api_url);
            if (message.sync_options != null && Object.hasOwnProperty.call(message, "sync_options"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.sync_options);
            if (message.linked_by_user_id != null && Object.hasOwnProperty.call(message, "linked_by_user_id"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.linked_by_user_id);
            if (message.linked_at != null && Object.hasOwnProperty.call(message, "linked_at"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.linked_at);
            if (message.enabled != null && Object.hasOwnProperty.call(message, "enabled"))
                writer.uint32(/* id 10, wireType 0 =*/80).bool(message.enabled);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 11, wireType 0 =*/88).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified IntegrationProject message, length delimited. Does not implicitly {@link repository.IntegrationProject.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.IntegrationProject
         * @static
         * @param {repository.IIntegrationProject} message IntegrationProject message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        IntegrationProject.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an IntegrationProject message from the specified reader or buffer.
         * @function decode
         * @memberof repository.IntegrationProject
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.IntegrationProject} IntegrationProject
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IntegrationProject.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.IntegrationProject();
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
                        message.integration_id = reader.string();
                        break;
                    }
                case 4: {
                        message.external_project_id = reader.string();
                        break;
                    }
                case 5: {
                        message.external_project_name = reader.string();
                        break;
                    }
                case 6: {
                        message.api_url = reader.string();
                        break;
                    }
                case 7: {
                        message.sync_options = reader.string();
                        break;
                    }
                case 8: {
                        message.linked_by_user_id = reader.string();
                        break;
                    }
                case 9: {
                        message.linked_at = reader.string();
                        break;
                    }
                case 10: {
                        message.enabled = reader.bool();
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
         * Decodes an IntegrationProject message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.IntegrationProject
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.IntegrationProject} IntegrationProject
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IntegrationProject.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an IntegrationProject message.
         * @function verify
         * @memberof repository.IntegrationProject
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        IntegrationProject.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.integration_id != null && message.hasOwnProperty("integration_id"))
                if (!$util.isString(message.integration_id))
                    return "integration_id: string expected";
            if (message.external_project_id != null && message.hasOwnProperty("external_project_id"))
                if (!$util.isString(message.external_project_id))
                    return "external_project_id: string expected";
            if (message.external_project_name != null && message.hasOwnProperty("external_project_name"))
                if (!$util.isString(message.external_project_name))
                    return "external_project_name: string expected";
            if (message.api_url != null && message.hasOwnProperty("api_url"))
                if (!$util.isString(message.api_url))
                    return "api_url: string expected";
            if (message.sync_options != null && message.hasOwnProperty("sync_options"))
                if (!$util.isString(message.sync_options))
                    return "sync_options: string expected";
            if (message.linked_by_user_id != null && message.hasOwnProperty("linked_by_user_id"))
                if (!$util.isString(message.linked_by_user_id))
                    return "linked_by_user_id: string expected";
            if (message.linked_at != null && message.hasOwnProperty("linked_at"))
                if (!$util.isString(message.linked_at))
                    return "linked_at: string expected";
            if (message.enabled != null && message.hasOwnProperty("enabled"))
                if (typeof message.enabled !== "boolean")
                    return "enabled: boolean expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates an IntegrationProject message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.IntegrationProject
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.IntegrationProject} IntegrationProject
         */
        IntegrationProject.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.IntegrationProject)
                return object;
            let message = new $root.repository.IntegrationProject();
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
            if (object.integration_id != null)
                message.integration_id = String(object.integration_id);
            if (object.external_project_id != null)
                message.external_project_id = String(object.external_project_id);
            if (object.external_project_name != null)
                message.external_project_name = String(object.external_project_name);
            if (object.api_url != null)
                message.api_url = String(object.api_url);
            if (object.sync_options != null)
                message.sync_options = String(object.sync_options);
            if (object.linked_by_user_id != null)
                message.linked_by_user_id = String(object.linked_by_user_id);
            if (object.linked_at != null)
                message.linked_at = String(object.linked_at);
            if (object.enabled != null)
                message.enabled = Boolean(object.enabled);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from an IntegrationProject message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.IntegrationProject
         * @static
         * @param {repository.IntegrationProject} message IntegrationProject
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        IntegrationProject.toObject = function toObject(message, options) {
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
                object.integration_id = "";
                object.external_project_id = "";
                object.external_project_name = "";
                object.api_url = "";
                object.sync_options = "";
                object.linked_by_user_id = "";
                object.linked_at = "";
                object.enabled = false;
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.integration_id != null && message.hasOwnProperty("integration_id"))
                object.integration_id = message.integration_id;
            if (message.external_project_id != null && message.hasOwnProperty("external_project_id"))
                object.external_project_id = message.external_project_id;
            if (message.external_project_name != null && message.hasOwnProperty("external_project_name"))
                object.external_project_name = message.external_project_name;
            if (message.api_url != null && message.hasOwnProperty("api_url"))
                object.api_url = message.api_url;
            if (message.sync_options != null && message.hasOwnProperty("sync_options"))
                object.sync_options = message.sync_options;
            if (message.linked_by_user_id != null && message.hasOwnProperty("linked_by_user_id"))
                object.linked_by_user_id = message.linked_by_user_id;
            if (message.linked_at != null && message.hasOwnProperty("linked_at"))
                object.linked_at = message.linked_at;
            if (message.enabled != null && message.hasOwnProperty("enabled"))
                object.enabled = message.enabled;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this IntegrationProject to JSON.
         * @function toJSON
         * @memberof repository.IntegrationProject
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        IntegrationProject.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for IntegrationProject
         * @function getTypeUrl
         * @memberof repository.IntegrationProject
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        IntegrationProject.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.IntegrationProject";
        };

        return IntegrationProject;
    })();

    repository.IntegrationCollectionMapping = (function() {

        /**
         * Properties of an IntegrationCollectionMapping.
         * @memberof repository
         * @interface IIntegrationCollectionMapping
         * @property {string|null} [id] IntegrationCollectionMapping id
         * @property {number|Long|null} [mtime] IntegrationCollectionMapping mtime
         * @property {string|null} [integration_id] IntegrationCollectionMapping integration_id
         * @property {string|null} [external_id] IntegrationCollectionMapping external_id
         * @property {string|null} [external_type] IntegrationCollectionMapping external_type
         * @property {string|null} [external_name] IntegrationCollectionMapping external_name
         * @property {string|null} [external_parent_id] IntegrationCollectionMapping external_parent_id
         * @property {string|null} [external_path] IntegrationCollectionMapping external_path
         * @property {string|null} [external_metadata] IntegrationCollectionMapping external_metadata
         * @property {string|null} [collection_id] IntegrationCollectionMapping collection_id
         * @property {string|null} [synced_at] IntegrationCollectionMapping synced_at
         * @property {boolean|null} [synced] IntegrationCollectionMapping synced
         */

        /**
         * Constructs a new IntegrationCollectionMapping.
         * @memberof repository
         * @classdesc Represents an IntegrationCollectionMapping.
         * @implements IIntegrationCollectionMapping
         * @constructor
         * @param {repository.IIntegrationCollectionMapping=} [properties] Properties to set
         */
        function IntegrationCollectionMapping(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * IntegrationCollectionMapping id.
         * @member {string} id
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.id = "";

        /**
         * IntegrationCollectionMapping mtime.
         * @member {number|Long} mtime
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * IntegrationCollectionMapping integration_id.
         * @member {string} integration_id
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.integration_id = "";

        /**
         * IntegrationCollectionMapping external_id.
         * @member {string} external_id
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.external_id = "";

        /**
         * IntegrationCollectionMapping external_type.
         * @member {string} external_type
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.external_type = "";

        /**
         * IntegrationCollectionMapping external_name.
         * @member {string} external_name
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.external_name = "";

        /**
         * IntegrationCollectionMapping external_parent_id.
         * @member {string} external_parent_id
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.external_parent_id = "";

        /**
         * IntegrationCollectionMapping external_path.
         * @member {string} external_path
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.external_path = "";

        /**
         * IntegrationCollectionMapping external_metadata.
         * @member {string} external_metadata
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.external_metadata = "";

        /**
         * IntegrationCollectionMapping collection_id.
         * @member {string} collection_id
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.collection_id = "";

        /**
         * IntegrationCollectionMapping synced_at.
         * @member {string} synced_at
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.synced_at = "";

        /**
         * IntegrationCollectionMapping synced.
         * @member {boolean} synced
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         */
        IntegrationCollectionMapping.prototype.synced = false;

        /**
         * Creates a new IntegrationCollectionMapping instance using the specified properties.
         * @function create
         * @memberof repository.IntegrationCollectionMapping
         * @static
         * @param {repository.IIntegrationCollectionMapping=} [properties] Properties to set
         * @returns {repository.IntegrationCollectionMapping} IntegrationCollectionMapping instance
         */
        IntegrationCollectionMapping.create = function create(properties) {
            return new IntegrationCollectionMapping(properties);
        };

        /**
         * Encodes the specified IntegrationCollectionMapping message. Does not implicitly {@link repository.IntegrationCollectionMapping.verify|verify} messages.
         * @function encode
         * @memberof repository.IntegrationCollectionMapping
         * @static
         * @param {repository.IIntegrationCollectionMapping} message IntegrationCollectionMapping message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        IntegrationCollectionMapping.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.integration_id != null && Object.hasOwnProperty.call(message, "integration_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.integration_id);
            if (message.external_id != null && Object.hasOwnProperty.call(message, "external_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.external_id);
            if (message.external_type != null && Object.hasOwnProperty.call(message, "external_type"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.external_type);
            if (message.external_name != null && Object.hasOwnProperty.call(message, "external_name"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.external_name);
            if (message.external_parent_id != null && Object.hasOwnProperty.call(message, "external_parent_id"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.external_parent_id);
            if (message.external_path != null && Object.hasOwnProperty.call(message, "external_path"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.external_path);
            if (message.external_metadata != null && Object.hasOwnProperty.call(message, "external_metadata"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.external_metadata);
            if (message.collection_id != null && Object.hasOwnProperty.call(message, "collection_id"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.collection_id);
            if (message.synced_at != null && Object.hasOwnProperty.call(message, "synced_at"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.synced_at);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 12, wireType 0 =*/96).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified IntegrationCollectionMapping message, length delimited. Does not implicitly {@link repository.IntegrationCollectionMapping.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.IntegrationCollectionMapping
         * @static
         * @param {repository.IIntegrationCollectionMapping} message IntegrationCollectionMapping message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        IntegrationCollectionMapping.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an IntegrationCollectionMapping message from the specified reader or buffer.
         * @function decode
         * @memberof repository.IntegrationCollectionMapping
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.IntegrationCollectionMapping} IntegrationCollectionMapping
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IntegrationCollectionMapping.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.IntegrationCollectionMapping();
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
                        message.integration_id = reader.string();
                        break;
                    }
                case 4: {
                        message.external_id = reader.string();
                        break;
                    }
                case 5: {
                        message.external_type = reader.string();
                        break;
                    }
                case 6: {
                        message.external_name = reader.string();
                        break;
                    }
                case 7: {
                        message.external_parent_id = reader.string();
                        break;
                    }
                case 8: {
                        message.external_path = reader.string();
                        break;
                    }
                case 9: {
                        message.external_metadata = reader.string();
                        break;
                    }
                case 10: {
                        message.collection_id = reader.string();
                        break;
                    }
                case 11: {
                        message.synced_at = reader.string();
                        break;
                    }
                case 12: {
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
         * Decodes an IntegrationCollectionMapping message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.IntegrationCollectionMapping
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.IntegrationCollectionMapping} IntegrationCollectionMapping
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IntegrationCollectionMapping.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an IntegrationCollectionMapping message.
         * @function verify
         * @memberof repository.IntegrationCollectionMapping
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        IntegrationCollectionMapping.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.integration_id != null && message.hasOwnProperty("integration_id"))
                if (!$util.isString(message.integration_id))
                    return "integration_id: string expected";
            if (message.external_id != null && message.hasOwnProperty("external_id"))
                if (!$util.isString(message.external_id))
                    return "external_id: string expected";
            if (message.external_type != null && message.hasOwnProperty("external_type"))
                if (!$util.isString(message.external_type))
                    return "external_type: string expected";
            if (message.external_name != null && message.hasOwnProperty("external_name"))
                if (!$util.isString(message.external_name))
                    return "external_name: string expected";
            if (message.external_parent_id != null && message.hasOwnProperty("external_parent_id"))
                if (!$util.isString(message.external_parent_id))
                    return "external_parent_id: string expected";
            if (message.external_path != null && message.hasOwnProperty("external_path"))
                if (!$util.isString(message.external_path))
                    return "external_path: string expected";
            if (message.external_metadata != null && message.hasOwnProperty("external_metadata"))
                if (!$util.isString(message.external_metadata))
                    return "external_metadata: string expected";
            if (message.collection_id != null && message.hasOwnProperty("collection_id"))
                if (!$util.isString(message.collection_id))
                    return "collection_id: string expected";
            if (message.synced_at != null && message.hasOwnProperty("synced_at"))
                if (!$util.isString(message.synced_at))
                    return "synced_at: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates an IntegrationCollectionMapping message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.IntegrationCollectionMapping
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.IntegrationCollectionMapping} IntegrationCollectionMapping
         */
        IntegrationCollectionMapping.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.IntegrationCollectionMapping)
                return object;
            let message = new $root.repository.IntegrationCollectionMapping();
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
            if (object.integration_id != null)
                message.integration_id = String(object.integration_id);
            if (object.external_id != null)
                message.external_id = String(object.external_id);
            if (object.external_type != null)
                message.external_type = String(object.external_type);
            if (object.external_name != null)
                message.external_name = String(object.external_name);
            if (object.external_parent_id != null)
                message.external_parent_id = String(object.external_parent_id);
            if (object.external_path != null)
                message.external_path = String(object.external_path);
            if (object.external_metadata != null)
                message.external_metadata = String(object.external_metadata);
            if (object.collection_id != null)
                message.collection_id = String(object.collection_id);
            if (object.synced_at != null)
                message.synced_at = String(object.synced_at);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from an IntegrationCollectionMapping message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.IntegrationCollectionMapping
         * @static
         * @param {repository.IntegrationCollectionMapping} message IntegrationCollectionMapping
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        IntegrationCollectionMapping.toObject = function toObject(message, options) {
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
                object.integration_id = "";
                object.external_id = "";
                object.external_type = "";
                object.external_name = "";
                object.external_parent_id = "";
                object.external_path = "";
                object.external_metadata = "";
                object.collection_id = "";
                object.synced_at = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.integration_id != null && message.hasOwnProperty("integration_id"))
                object.integration_id = message.integration_id;
            if (message.external_id != null && message.hasOwnProperty("external_id"))
                object.external_id = message.external_id;
            if (message.external_type != null && message.hasOwnProperty("external_type"))
                object.external_type = message.external_type;
            if (message.external_name != null && message.hasOwnProperty("external_name"))
                object.external_name = message.external_name;
            if (message.external_parent_id != null && message.hasOwnProperty("external_parent_id"))
                object.external_parent_id = message.external_parent_id;
            if (message.external_path != null && message.hasOwnProperty("external_path"))
                object.external_path = message.external_path;
            if (message.external_metadata != null && message.hasOwnProperty("external_metadata"))
                object.external_metadata = message.external_metadata;
            if (message.collection_id != null && message.hasOwnProperty("collection_id"))
                object.collection_id = message.collection_id;
            if (message.synced_at != null && message.hasOwnProperty("synced_at"))
                object.synced_at = message.synced_at;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this IntegrationCollectionMapping to JSON.
         * @function toJSON
         * @memberof repository.IntegrationCollectionMapping
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        IntegrationCollectionMapping.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for IntegrationCollectionMapping
         * @function getTypeUrl
         * @memberof repository.IntegrationCollectionMapping
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        IntegrationCollectionMapping.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.IntegrationCollectionMapping";
        };

        return IntegrationCollectionMapping;
    })();

    repository.IntegrationAssetMapping = (function() {

        /**
         * Properties of an IntegrationAssetMapping.
         * @memberof repository
         * @interface IIntegrationAssetMapping
         * @property {string|null} [id] IntegrationAssetMapping id
         * @property {number|Long|null} [mtime] IntegrationAssetMapping mtime
         * @property {string|null} [integration_id] IntegrationAssetMapping integration_id
         * @property {string|null} [external_id] IntegrationAssetMapping external_id
         * @property {string|null} [external_name] IntegrationAssetMapping external_name
         * @property {string|null} [external_parent_id] IntegrationAssetMapping external_parent_id
         * @property {string|null} [external_type] IntegrationAssetMapping external_type
         * @property {string|null} [external_status] IntegrationAssetMapping external_status
         * @property {string|null} [external_assignees] IntegrationAssetMapping external_assignees
         * @property {string|null} [external_metadata] IntegrationAssetMapping external_metadata
         * @property {string|null} [asset_id] IntegrationAssetMapping asset_id
         * @property {string|null} [last_pushed_checkpoint_id] IntegrationAssetMapping last_pushed_checkpoint_id
         * @property {string|null} [synced_at] IntegrationAssetMapping synced_at
         * @property {boolean|null} [synced] IntegrationAssetMapping synced
         */

        /**
         * Constructs a new IntegrationAssetMapping.
         * @memberof repository
         * @classdesc Represents an IntegrationAssetMapping.
         * @implements IIntegrationAssetMapping
         * @constructor
         * @param {repository.IIntegrationAssetMapping=} [properties] Properties to set
         */
        function IntegrationAssetMapping(properties) {
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * IntegrationAssetMapping id.
         * @member {string} id
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.id = "";

        /**
         * IntegrationAssetMapping mtime.
         * @member {number|Long} mtime
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * IntegrationAssetMapping integration_id.
         * @member {string} integration_id
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.integration_id = "";

        /**
         * IntegrationAssetMapping external_id.
         * @member {string} external_id
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.external_id = "";

        /**
         * IntegrationAssetMapping external_name.
         * @member {string} external_name
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.external_name = "";

        /**
         * IntegrationAssetMapping external_parent_id.
         * @member {string} external_parent_id
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.external_parent_id = "";

        /**
         * IntegrationAssetMapping external_type.
         * @member {string} external_type
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.external_type = "";

        /**
         * IntegrationAssetMapping external_status.
         * @member {string} external_status
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.external_status = "";

        /**
         * IntegrationAssetMapping external_assignees.
         * @member {string} external_assignees
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.external_assignees = "";

        /**
         * IntegrationAssetMapping external_metadata.
         * @member {string} external_metadata
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.external_metadata = "";

        /**
         * IntegrationAssetMapping asset_id.
         * @member {string} asset_id
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.asset_id = "";

        /**
         * IntegrationAssetMapping last_pushed_checkpoint_id.
         * @member {string} last_pushed_checkpoint_id
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.last_pushed_checkpoint_id = "";

        /**
         * IntegrationAssetMapping synced_at.
         * @member {string} synced_at
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.synced_at = "";

        /**
         * IntegrationAssetMapping synced.
         * @member {boolean} synced
         * @memberof repository.IntegrationAssetMapping
         * @instance
         */
        IntegrationAssetMapping.prototype.synced = false;

        /**
         * Creates a new IntegrationAssetMapping instance using the specified properties.
         * @function create
         * @memberof repository.IntegrationAssetMapping
         * @static
         * @param {repository.IIntegrationAssetMapping=} [properties] Properties to set
         * @returns {repository.IntegrationAssetMapping} IntegrationAssetMapping instance
         */
        IntegrationAssetMapping.create = function create(properties) {
            return new IntegrationAssetMapping(properties);
        };

        /**
         * Encodes the specified IntegrationAssetMapping message. Does not implicitly {@link repository.IntegrationAssetMapping.verify|verify} messages.
         * @function encode
         * @memberof repository.IntegrationAssetMapping
         * @static
         * @param {repository.IIntegrationAssetMapping} message IntegrationAssetMapping message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        IntegrationAssetMapping.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.id != null && Object.hasOwnProperty.call(message, "id"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.id);
            if (message.mtime != null && Object.hasOwnProperty.call(message, "mtime"))
                writer.uint32(/* id 2, wireType 0 =*/16).int64(message.mtime);
            if (message.integration_id != null && Object.hasOwnProperty.call(message, "integration_id"))
                writer.uint32(/* id 3, wireType 2 =*/26).string(message.integration_id);
            if (message.external_id != null && Object.hasOwnProperty.call(message, "external_id"))
                writer.uint32(/* id 4, wireType 2 =*/34).string(message.external_id);
            if (message.external_name != null && Object.hasOwnProperty.call(message, "external_name"))
                writer.uint32(/* id 5, wireType 2 =*/42).string(message.external_name);
            if (message.external_parent_id != null && Object.hasOwnProperty.call(message, "external_parent_id"))
                writer.uint32(/* id 6, wireType 2 =*/50).string(message.external_parent_id);
            if (message.external_type != null && Object.hasOwnProperty.call(message, "external_type"))
                writer.uint32(/* id 7, wireType 2 =*/58).string(message.external_type);
            if (message.external_status != null && Object.hasOwnProperty.call(message, "external_status"))
                writer.uint32(/* id 8, wireType 2 =*/66).string(message.external_status);
            if (message.external_assignees != null && Object.hasOwnProperty.call(message, "external_assignees"))
                writer.uint32(/* id 9, wireType 2 =*/74).string(message.external_assignees);
            if (message.external_metadata != null && Object.hasOwnProperty.call(message, "external_metadata"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.external_metadata);
            if (message.asset_id != null && Object.hasOwnProperty.call(message, "asset_id"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.asset_id);
            if (message.last_pushed_checkpoint_id != null && Object.hasOwnProperty.call(message, "last_pushed_checkpoint_id"))
                writer.uint32(/* id 12, wireType 2 =*/98).string(message.last_pushed_checkpoint_id);
            if (message.synced_at != null && Object.hasOwnProperty.call(message, "synced_at"))
                writer.uint32(/* id 13, wireType 2 =*/106).string(message.synced_at);
            if (message.synced != null && Object.hasOwnProperty.call(message, "synced"))
                writer.uint32(/* id 14, wireType 0 =*/112).bool(message.synced);
            return writer;
        };

        /**
         * Encodes the specified IntegrationAssetMapping message, length delimited. Does not implicitly {@link repository.IntegrationAssetMapping.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.IntegrationAssetMapping
         * @static
         * @param {repository.IIntegrationAssetMapping} message IntegrationAssetMapping message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        IntegrationAssetMapping.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an IntegrationAssetMapping message from the specified reader or buffer.
         * @function decode
         * @memberof repository.IntegrationAssetMapping
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.IntegrationAssetMapping} IntegrationAssetMapping
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IntegrationAssetMapping.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.IntegrationAssetMapping();
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
                        message.integration_id = reader.string();
                        break;
                    }
                case 4: {
                        message.external_id = reader.string();
                        break;
                    }
                case 5: {
                        message.external_name = reader.string();
                        break;
                    }
                case 6: {
                        message.external_parent_id = reader.string();
                        break;
                    }
                case 7: {
                        message.external_type = reader.string();
                        break;
                    }
                case 8: {
                        message.external_status = reader.string();
                        break;
                    }
                case 9: {
                        message.external_assignees = reader.string();
                        break;
                    }
                case 10: {
                        message.external_metadata = reader.string();
                        break;
                    }
                case 11: {
                        message.asset_id = reader.string();
                        break;
                    }
                case 12: {
                        message.last_pushed_checkpoint_id = reader.string();
                        break;
                    }
                case 13: {
                        message.synced_at = reader.string();
                        break;
                    }
                case 14: {
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
         * Decodes an IntegrationAssetMapping message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.IntegrationAssetMapping
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.IntegrationAssetMapping} IntegrationAssetMapping
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        IntegrationAssetMapping.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an IntegrationAssetMapping message.
         * @function verify
         * @memberof repository.IntegrationAssetMapping
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        IntegrationAssetMapping.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.id != null && message.hasOwnProperty("id"))
                if (!$util.isString(message.id))
                    return "id: string expected";
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (!$util.isInteger(message.mtime) && !(message.mtime && $util.isInteger(message.mtime.low) && $util.isInteger(message.mtime.high)))
                    return "mtime: integer|Long expected";
            if (message.integration_id != null && message.hasOwnProperty("integration_id"))
                if (!$util.isString(message.integration_id))
                    return "integration_id: string expected";
            if (message.external_id != null && message.hasOwnProperty("external_id"))
                if (!$util.isString(message.external_id))
                    return "external_id: string expected";
            if (message.external_name != null && message.hasOwnProperty("external_name"))
                if (!$util.isString(message.external_name))
                    return "external_name: string expected";
            if (message.external_parent_id != null && message.hasOwnProperty("external_parent_id"))
                if (!$util.isString(message.external_parent_id))
                    return "external_parent_id: string expected";
            if (message.external_type != null && message.hasOwnProperty("external_type"))
                if (!$util.isString(message.external_type))
                    return "external_type: string expected";
            if (message.external_status != null && message.hasOwnProperty("external_status"))
                if (!$util.isString(message.external_status))
                    return "external_status: string expected";
            if (message.external_assignees != null && message.hasOwnProperty("external_assignees"))
                if (!$util.isString(message.external_assignees))
                    return "external_assignees: string expected";
            if (message.external_metadata != null && message.hasOwnProperty("external_metadata"))
                if (!$util.isString(message.external_metadata))
                    return "external_metadata: string expected";
            if (message.asset_id != null && message.hasOwnProperty("asset_id"))
                if (!$util.isString(message.asset_id))
                    return "asset_id: string expected";
            if (message.last_pushed_checkpoint_id != null && message.hasOwnProperty("last_pushed_checkpoint_id"))
                if (!$util.isString(message.last_pushed_checkpoint_id))
                    return "last_pushed_checkpoint_id: string expected";
            if (message.synced_at != null && message.hasOwnProperty("synced_at"))
                if (!$util.isString(message.synced_at))
                    return "synced_at: string expected";
            if (message.synced != null && message.hasOwnProperty("synced"))
                if (typeof message.synced !== "boolean")
                    return "synced: boolean expected";
            return null;
        };

        /**
         * Creates an IntegrationAssetMapping message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.IntegrationAssetMapping
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.IntegrationAssetMapping} IntegrationAssetMapping
         */
        IntegrationAssetMapping.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.IntegrationAssetMapping)
                return object;
            let message = new $root.repository.IntegrationAssetMapping();
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
            if (object.integration_id != null)
                message.integration_id = String(object.integration_id);
            if (object.external_id != null)
                message.external_id = String(object.external_id);
            if (object.external_name != null)
                message.external_name = String(object.external_name);
            if (object.external_parent_id != null)
                message.external_parent_id = String(object.external_parent_id);
            if (object.external_type != null)
                message.external_type = String(object.external_type);
            if (object.external_status != null)
                message.external_status = String(object.external_status);
            if (object.external_assignees != null)
                message.external_assignees = String(object.external_assignees);
            if (object.external_metadata != null)
                message.external_metadata = String(object.external_metadata);
            if (object.asset_id != null)
                message.asset_id = String(object.asset_id);
            if (object.last_pushed_checkpoint_id != null)
                message.last_pushed_checkpoint_id = String(object.last_pushed_checkpoint_id);
            if (object.synced_at != null)
                message.synced_at = String(object.synced_at);
            if (object.synced != null)
                message.synced = Boolean(object.synced);
            return message;
        };

        /**
         * Creates a plain object from an IntegrationAssetMapping message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.IntegrationAssetMapping
         * @static
         * @param {repository.IntegrationAssetMapping} message IntegrationAssetMapping
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        IntegrationAssetMapping.toObject = function toObject(message, options) {
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
                object.integration_id = "";
                object.external_id = "";
                object.external_name = "";
                object.external_parent_id = "";
                object.external_type = "";
                object.external_status = "";
                object.external_assignees = "";
                object.external_metadata = "";
                object.asset_id = "";
                object.last_pushed_checkpoint_id = "";
                object.synced_at = "";
                object.synced = false;
            }
            if (message.id != null && message.hasOwnProperty("id"))
                object.id = message.id;
            if (message.mtime != null && message.hasOwnProperty("mtime"))
                if (typeof message.mtime === "number")
                    object.mtime = options.longs === String ? String(message.mtime) : message.mtime;
                else
                    object.mtime = options.longs === String ? $util.Long.prototype.toString.call(message.mtime) : options.longs === Number ? new $util.LongBits(message.mtime.low >>> 0, message.mtime.high >>> 0).toNumber() : message.mtime;
            if (message.integration_id != null && message.hasOwnProperty("integration_id"))
                object.integration_id = message.integration_id;
            if (message.external_id != null && message.hasOwnProperty("external_id"))
                object.external_id = message.external_id;
            if (message.external_name != null && message.hasOwnProperty("external_name"))
                object.external_name = message.external_name;
            if (message.external_parent_id != null && message.hasOwnProperty("external_parent_id"))
                object.external_parent_id = message.external_parent_id;
            if (message.external_type != null && message.hasOwnProperty("external_type"))
                object.external_type = message.external_type;
            if (message.external_status != null && message.hasOwnProperty("external_status"))
                object.external_status = message.external_status;
            if (message.external_assignees != null && message.hasOwnProperty("external_assignees"))
                object.external_assignees = message.external_assignees;
            if (message.external_metadata != null && message.hasOwnProperty("external_metadata"))
                object.external_metadata = message.external_metadata;
            if (message.asset_id != null && message.hasOwnProperty("asset_id"))
                object.asset_id = message.asset_id;
            if (message.last_pushed_checkpoint_id != null && message.hasOwnProperty("last_pushed_checkpoint_id"))
                object.last_pushed_checkpoint_id = message.last_pushed_checkpoint_id;
            if (message.synced_at != null && message.hasOwnProperty("synced_at"))
                object.synced_at = message.synced_at;
            if (message.synced != null && message.hasOwnProperty("synced"))
                object.synced = message.synced;
            return object;
        };

        /**
         * Converts this IntegrationAssetMapping to JSON.
         * @function toJSON
         * @memberof repository.IntegrationAssetMapping
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        IntegrationAssetMapping.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for IntegrationAssetMapping
         * @function getTypeUrl
         * @memberof repository.IntegrationAssetMapping
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        IntegrationAssetMapping.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.IntegrationAssetMapping";
        };

        return IntegrationAssetMapping;
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
         * @property {Array.<repository.IAsset>|null} [assets] ProjectData assets
         * @property {Array.<repository.IAssetType>|null} [asset_types] ProjectData asset_types
         * @property {Array.<repository.ICheckpoint>|null} [asset_checkpoints] ProjectData asset_checkpoints
         * @property {Array.<repository.IAssetDependency>|null} [asset_dependencies] ProjectData asset_dependencies
         * @property {Array.<repository.ICollectionDependency>|null} [collection_dependencies] ProjectData collection_dependencies
         * @property {Array.<repository.IStatus>|null} [statuses] ProjectData statuses
         * @property {Array.<repository.IDependencyType>|null} [dependency_types] ProjectData dependency_types
         * @property {Array.<repository.IUser>|null} [users] ProjectData users
         * @property {Array.<repository.IRole>|null} [roles] ProjectData roles
         * @property {Array.<repository.ICollectionType>|null} [collection_types] ProjectData collection_types
         * @property {Array.<repository.ICollection>|null} [collections] ProjectData collections
         * @property {Array.<repository.ICollectionAssignee>|null} [collection_assignees] ProjectData collection_assignees
         * @property {Array.<repository.ITemplate>|null} [templates] ProjectData templates
         * @property {Array.<repository.ITag>|null} [tags] ProjectData tags
         * @property {Array.<repository.IAssetTag>|null} [asset_tags] ProjectData asset_tags
         * @property {Array.<repository.IWorkflow>|null} [workflows] ProjectData workflows
         * @property {Array.<repository.IWorkflowLink>|null} [workflow_links] ProjectData workflow_links
         * @property {Array.<repository.IWorkflowCollection>|null} [workflow_collections] ProjectData workflow_collections
         * @property {Array.<repository.IWorkflowAsset>|null} [workflow_assets] ProjectData workflow_assets
         * @property {Array.<repository.ITomb>|null} [tomb] ProjectData tomb
         * @property {Array.<repository.IIntegrationProject>|null} [integration_projects] ProjectData integration_projects
         * @property {Array.<repository.IIntegrationCollectionMapping>|null} [integration_collection_mappings] ProjectData integration_collection_mappings
         * @property {Array.<repository.IIntegrationAssetMapping>|null} [integration_asset_mappings] ProjectData integration_asset_mappings
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
            this.assets = [];
            this.asset_types = [];
            this.asset_checkpoints = [];
            this.asset_dependencies = [];
            this.collection_dependencies = [];
            this.statuses = [];
            this.dependency_types = [];
            this.users = [];
            this.roles = [];
            this.collection_types = [];
            this.collections = [];
            this.collection_assignees = [];
            this.templates = [];
            this.tags = [];
            this.asset_tags = [];
            this.workflows = [];
            this.workflow_links = [];
            this.workflow_collections = [];
            this.workflow_assets = [];
            this.tomb = [];
            this.integration_projects = [];
            this.integration_collection_mappings = [];
            this.integration_asset_mappings = [];
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
         * ProjectData assets.
         * @member {Array.<repository.IAsset>} assets
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.assets = $util.emptyArray;

        /**
         * ProjectData asset_types.
         * @member {Array.<repository.IAssetType>} asset_types
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.asset_types = $util.emptyArray;

        /**
         * ProjectData asset_checkpoints.
         * @member {Array.<repository.ICheckpoint>} asset_checkpoints
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.asset_checkpoints = $util.emptyArray;

        /**
         * ProjectData asset_dependencies.
         * @member {Array.<repository.IAssetDependency>} asset_dependencies
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.asset_dependencies = $util.emptyArray;

        /**
         * ProjectData collection_dependencies.
         * @member {Array.<repository.ICollectionDependency>} collection_dependencies
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.collection_dependencies = $util.emptyArray;

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
         * ProjectData collection_types.
         * @member {Array.<repository.ICollectionType>} collection_types
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.collection_types = $util.emptyArray;

        /**
         * ProjectData collections.
         * @member {Array.<repository.ICollection>} collections
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.collections = $util.emptyArray;

        /**
         * ProjectData collection_assignees.
         * @member {Array.<repository.ICollectionAssignee>} collection_assignees
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.collection_assignees = $util.emptyArray;

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
         * ProjectData asset_tags.
         * @member {Array.<repository.IAssetTag>} asset_tags
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.asset_tags = $util.emptyArray;

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
         * ProjectData workflow_collections.
         * @member {Array.<repository.IWorkflowCollection>} workflow_collections
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.workflow_collections = $util.emptyArray;

        /**
         * ProjectData workflow_assets.
         * @member {Array.<repository.IWorkflowAsset>} workflow_assets
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.workflow_assets = $util.emptyArray;

        /**
         * ProjectData tomb.
         * @member {Array.<repository.ITomb>} tomb
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.tomb = $util.emptyArray;

        /**
         * ProjectData integration_projects.
         * @member {Array.<repository.IIntegrationProject>} integration_projects
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.integration_projects = $util.emptyArray;

        /**
         * ProjectData integration_collection_mappings.
         * @member {Array.<repository.IIntegrationCollectionMapping>} integration_collection_mappings
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.integration_collection_mappings = $util.emptyArray;

        /**
         * ProjectData integration_asset_mappings.
         * @member {Array.<repository.IIntegrationAssetMapping>} integration_asset_mappings
         * @memberof repository.ProjectData
         * @instance
         */
        ProjectData.prototype.integration_asset_mappings = $util.emptyArray;

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
            if (message.assets != null && message.assets.length)
                for (let i = 0; i < message.assets.length; ++i)
                    $root.repository.Asset.encode(message.assets[i], writer.uint32(/* id 2, wireType 2 =*/18).fork()).ldelim();
            if (message.asset_types != null && message.asset_types.length)
                for (let i = 0; i < message.asset_types.length; ++i)
                    $root.repository.AssetType.encode(message.asset_types[i], writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            if (message.asset_checkpoints != null && message.asset_checkpoints.length)
                for (let i = 0; i < message.asset_checkpoints.length; ++i)
                    $root.repository.Checkpoint.encode(message.asset_checkpoints[i], writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
            if (message.asset_dependencies != null && message.asset_dependencies.length)
                for (let i = 0; i < message.asset_dependencies.length; ++i)
                    $root.repository.AssetDependency.encode(message.asset_dependencies[i], writer.uint32(/* id 5, wireType 2 =*/42).fork()).ldelim();
            if (message.collection_dependencies != null && message.collection_dependencies.length)
                for (let i = 0; i < message.collection_dependencies.length; ++i)
                    $root.repository.CollectionDependency.encode(message.collection_dependencies[i], writer.uint32(/* id 6, wireType 2 =*/50).fork()).ldelim();
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
            if (message.collection_types != null && message.collection_types.length)
                for (let i = 0; i < message.collection_types.length; ++i)
                    $root.repository.CollectionType.encode(message.collection_types[i], writer.uint32(/* id 11, wireType 2 =*/90).fork()).ldelim();
            if (message.collections != null && message.collections.length)
                for (let i = 0; i < message.collections.length; ++i)
                    $root.repository.Collection.encode(message.collections[i], writer.uint32(/* id 12, wireType 2 =*/98).fork()).ldelim();
            if (message.collection_assignees != null && message.collection_assignees.length)
                for (let i = 0; i < message.collection_assignees.length; ++i)
                    $root.repository.CollectionAssignee.encode(message.collection_assignees[i], writer.uint32(/* id 13, wireType 2 =*/106).fork()).ldelim();
            if (message.templates != null && message.templates.length)
                for (let i = 0; i < message.templates.length; ++i)
                    $root.repository.Template.encode(message.templates[i], writer.uint32(/* id 14, wireType 2 =*/114).fork()).ldelim();
            if (message.tags != null && message.tags.length)
                for (let i = 0; i < message.tags.length; ++i)
                    $root.repository.Tag.encode(message.tags[i], writer.uint32(/* id 15, wireType 2 =*/122).fork()).ldelim();
            if (message.asset_tags != null && message.asset_tags.length)
                for (let i = 0; i < message.asset_tags.length; ++i)
                    $root.repository.AssetTag.encode(message.asset_tags[i], writer.uint32(/* id 16, wireType 2 =*/130).fork()).ldelim();
            if (message.workflows != null && message.workflows.length)
                for (let i = 0; i < message.workflows.length; ++i)
                    $root.repository.Workflow.encode(message.workflows[i], writer.uint32(/* id 17, wireType 2 =*/138).fork()).ldelim();
            if (message.workflow_links != null && message.workflow_links.length)
                for (let i = 0; i < message.workflow_links.length; ++i)
                    $root.repository.WorkflowLink.encode(message.workflow_links[i], writer.uint32(/* id 18, wireType 2 =*/146).fork()).ldelim();
            if (message.workflow_collections != null && message.workflow_collections.length)
                for (let i = 0; i < message.workflow_collections.length; ++i)
                    $root.repository.WorkflowCollection.encode(message.workflow_collections[i], writer.uint32(/* id 19, wireType 2 =*/154).fork()).ldelim();
            if (message.workflow_assets != null && message.workflow_assets.length)
                for (let i = 0; i < message.workflow_assets.length; ++i)
                    $root.repository.WorkflowAsset.encode(message.workflow_assets[i], writer.uint32(/* id 20, wireType 2 =*/162).fork()).ldelim();
            if (message.tomb != null && message.tomb.length)
                for (let i = 0; i < message.tomb.length; ++i)
                    $root.repository.Tomb.encode(message.tomb[i], writer.uint32(/* id 21, wireType 2 =*/170).fork()).ldelim();
            if (message.integration_projects != null && message.integration_projects.length)
                for (let i = 0; i < message.integration_projects.length; ++i)
                    $root.repository.IntegrationProject.encode(message.integration_projects[i], writer.uint32(/* id 22, wireType 2 =*/178).fork()).ldelim();
            if (message.integration_collection_mappings != null && message.integration_collection_mappings.length)
                for (let i = 0; i < message.integration_collection_mappings.length; ++i)
                    $root.repository.IntegrationCollectionMapping.encode(message.integration_collection_mappings[i], writer.uint32(/* id 23, wireType 2 =*/186).fork()).ldelim();
            if (message.integration_asset_mappings != null && message.integration_asset_mappings.length)
                for (let i = 0; i < message.integration_asset_mappings.length; ++i)
                    $root.repository.IntegrationAssetMapping.encode(message.integration_asset_mappings[i], writer.uint32(/* id 24, wireType 2 =*/194).fork()).ldelim();
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
                        if (!(message.assets && message.assets.length))
                            message.assets = [];
                        message.assets.push($root.repository.Asset.decode(reader, reader.uint32()));
                        break;
                    }
                case 3: {
                        if (!(message.asset_types && message.asset_types.length))
                            message.asset_types = [];
                        message.asset_types.push($root.repository.AssetType.decode(reader, reader.uint32()));
                        break;
                    }
                case 4: {
                        if (!(message.asset_checkpoints && message.asset_checkpoints.length))
                            message.asset_checkpoints = [];
                        message.asset_checkpoints.push($root.repository.Checkpoint.decode(reader, reader.uint32()));
                        break;
                    }
                case 5: {
                        if (!(message.asset_dependencies && message.asset_dependencies.length))
                            message.asset_dependencies = [];
                        message.asset_dependencies.push($root.repository.AssetDependency.decode(reader, reader.uint32()));
                        break;
                    }
                case 6: {
                        if (!(message.collection_dependencies && message.collection_dependencies.length))
                            message.collection_dependencies = [];
                        message.collection_dependencies.push($root.repository.CollectionDependency.decode(reader, reader.uint32()));
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
                        if (!(message.collection_types && message.collection_types.length))
                            message.collection_types = [];
                        message.collection_types.push($root.repository.CollectionType.decode(reader, reader.uint32()));
                        break;
                    }
                case 12: {
                        if (!(message.collections && message.collections.length))
                            message.collections = [];
                        message.collections.push($root.repository.Collection.decode(reader, reader.uint32()));
                        break;
                    }
                case 13: {
                        if (!(message.collection_assignees && message.collection_assignees.length))
                            message.collection_assignees = [];
                        message.collection_assignees.push($root.repository.CollectionAssignee.decode(reader, reader.uint32()));
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
                        if (!(message.asset_tags && message.asset_tags.length))
                            message.asset_tags = [];
                        message.asset_tags.push($root.repository.AssetTag.decode(reader, reader.uint32()));
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
                        if (!(message.workflow_collections && message.workflow_collections.length))
                            message.workflow_collections = [];
                        message.workflow_collections.push($root.repository.WorkflowCollection.decode(reader, reader.uint32()));
                        break;
                    }
                case 20: {
                        if (!(message.workflow_assets && message.workflow_assets.length))
                            message.workflow_assets = [];
                        message.workflow_assets.push($root.repository.WorkflowAsset.decode(reader, reader.uint32()));
                        break;
                    }
                case 21: {
                        if (!(message.tomb && message.tomb.length))
                            message.tomb = [];
                        message.tomb.push($root.repository.Tomb.decode(reader, reader.uint32()));
                        break;
                    }
                case 22: {
                        if (!(message.integration_projects && message.integration_projects.length))
                            message.integration_projects = [];
                        message.integration_projects.push($root.repository.IntegrationProject.decode(reader, reader.uint32()));
                        break;
                    }
                case 23: {
                        if (!(message.integration_collection_mappings && message.integration_collection_mappings.length))
                            message.integration_collection_mappings = [];
                        message.integration_collection_mappings.push($root.repository.IntegrationCollectionMapping.decode(reader, reader.uint32()));
                        break;
                    }
                case 24: {
                        if (!(message.integration_asset_mappings && message.integration_asset_mappings.length))
                            message.integration_asset_mappings = [];
                        message.integration_asset_mappings.push($root.repository.IntegrationAssetMapping.decode(reader, reader.uint32()));
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
            if (message.assets != null && message.hasOwnProperty("assets")) {
                if (!Array.isArray(message.assets))
                    return "assets: array expected";
                for (let i = 0; i < message.assets.length; ++i) {
                    let error = $root.repository.Asset.verify(message.assets[i]);
                    if (error)
                        return "assets." + error;
                }
            }
            if (message.asset_types != null && message.hasOwnProperty("asset_types")) {
                if (!Array.isArray(message.asset_types))
                    return "asset_types: array expected";
                for (let i = 0; i < message.asset_types.length; ++i) {
                    let error = $root.repository.AssetType.verify(message.asset_types[i]);
                    if (error)
                        return "asset_types." + error;
                }
            }
            if (message.asset_checkpoints != null && message.hasOwnProperty("asset_checkpoints")) {
                if (!Array.isArray(message.asset_checkpoints))
                    return "asset_checkpoints: array expected";
                for (let i = 0; i < message.asset_checkpoints.length; ++i) {
                    let error = $root.repository.Checkpoint.verify(message.asset_checkpoints[i]);
                    if (error)
                        return "asset_checkpoints." + error;
                }
            }
            if (message.asset_dependencies != null && message.hasOwnProperty("asset_dependencies")) {
                if (!Array.isArray(message.asset_dependencies))
                    return "asset_dependencies: array expected";
                for (let i = 0; i < message.asset_dependencies.length; ++i) {
                    let error = $root.repository.AssetDependency.verify(message.asset_dependencies[i]);
                    if (error)
                        return "asset_dependencies." + error;
                }
            }
            if (message.collection_dependencies != null && message.hasOwnProperty("collection_dependencies")) {
                if (!Array.isArray(message.collection_dependencies))
                    return "collection_dependencies: array expected";
                for (let i = 0; i < message.collection_dependencies.length; ++i) {
                    let error = $root.repository.CollectionDependency.verify(message.collection_dependencies[i]);
                    if (error)
                        return "collection_dependencies." + error;
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
            if (message.collection_types != null && message.hasOwnProperty("collection_types")) {
                if (!Array.isArray(message.collection_types))
                    return "collection_types: array expected";
                for (let i = 0; i < message.collection_types.length; ++i) {
                    let error = $root.repository.CollectionType.verify(message.collection_types[i]);
                    if (error)
                        return "collection_types." + error;
                }
            }
            if (message.collections != null && message.hasOwnProperty("collections")) {
                if (!Array.isArray(message.collections))
                    return "collections: array expected";
                for (let i = 0; i < message.collections.length; ++i) {
                    let error = $root.repository.Collection.verify(message.collections[i]);
                    if (error)
                        return "collections." + error;
                }
            }
            if (message.collection_assignees != null && message.hasOwnProperty("collection_assignees")) {
                if (!Array.isArray(message.collection_assignees))
                    return "collection_assignees: array expected";
                for (let i = 0; i < message.collection_assignees.length; ++i) {
                    let error = $root.repository.CollectionAssignee.verify(message.collection_assignees[i]);
                    if (error)
                        return "collection_assignees." + error;
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
            if (message.asset_tags != null && message.hasOwnProperty("asset_tags")) {
                if (!Array.isArray(message.asset_tags))
                    return "asset_tags: array expected";
                for (let i = 0; i < message.asset_tags.length; ++i) {
                    let error = $root.repository.AssetTag.verify(message.asset_tags[i]);
                    if (error)
                        return "asset_tags." + error;
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
            if (message.workflow_collections != null && message.hasOwnProperty("workflow_collections")) {
                if (!Array.isArray(message.workflow_collections))
                    return "workflow_collections: array expected";
                for (let i = 0; i < message.workflow_collections.length; ++i) {
                    let error = $root.repository.WorkflowCollection.verify(message.workflow_collections[i]);
                    if (error)
                        return "workflow_collections." + error;
                }
            }
            if (message.workflow_assets != null && message.hasOwnProperty("workflow_assets")) {
                if (!Array.isArray(message.workflow_assets))
                    return "workflow_assets: array expected";
                for (let i = 0; i < message.workflow_assets.length; ++i) {
                    let error = $root.repository.WorkflowAsset.verify(message.workflow_assets[i]);
                    if (error)
                        return "workflow_assets." + error;
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
            if (message.integration_projects != null && message.hasOwnProperty("integration_projects")) {
                if (!Array.isArray(message.integration_projects))
                    return "integration_projects: array expected";
                for (let i = 0; i < message.integration_projects.length; ++i) {
                    let error = $root.repository.IntegrationProject.verify(message.integration_projects[i]);
                    if (error)
                        return "integration_projects." + error;
                }
            }
            if (message.integration_collection_mappings != null && message.hasOwnProperty("integration_collection_mappings")) {
                if (!Array.isArray(message.integration_collection_mappings))
                    return "integration_collection_mappings: array expected";
                for (let i = 0; i < message.integration_collection_mappings.length; ++i) {
                    let error = $root.repository.IntegrationCollectionMapping.verify(message.integration_collection_mappings[i]);
                    if (error)
                        return "integration_collection_mappings." + error;
                }
            }
            if (message.integration_asset_mappings != null && message.hasOwnProperty("integration_asset_mappings")) {
                if (!Array.isArray(message.integration_asset_mappings))
                    return "integration_asset_mappings: array expected";
                for (let i = 0; i < message.integration_asset_mappings.length; ++i) {
                    let error = $root.repository.IntegrationAssetMapping.verify(message.integration_asset_mappings[i]);
                    if (error)
                        return "integration_asset_mappings." + error;
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
            if (object.assets) {
                if (!Array.isArray(object.assets))
                    throw TypeError(".repository.ProjectData.assets: array expected");
                message.assets = [];
                for (let i = 0; i < object.assets.length; ++i) {
                    if (typeof object.assets[i] !== "object")
                        throw TypeError(".repository.ProjectData.assets: object expected");
                    message.assets[i] = $root.repository.Asset.fromObject(object.assets[i]);
                }
            }
            if (object.asset_types) {
                if (!Array.isArray(object.asset_types))
                    throw TypeError(".repository.ProjectData.asset_types: array expected");
                message.asset_types = [];
                for (let i = 0; i < object.asset_types.length; ++i) {
                    if (typeof object.asset_types[i] !== "object")
                        throw TypeError(".repository.ProjectData.asset_types: object expected");
                    message.asset_types[i] = $root.repository.AssetType.fromObject(object.asset_types[i]);
                }
            }
            if (object.asset_checkpoints) {
                if (!Array.isArray(object.asset_checkpoints))
                    throw TypeError(".repository.ProjectData.asset_checkpoints: array expected");
                message.asset_checkpoints = [];
                for (let i = 0; i < object.asset_checkpoints.length; ++i) {
                    if (typeof object.asset_checkpoints[i] !== "object")
                        throw TypeError(".repository.ProjectData.asset_checkpoints: object expected");
                    message.asset_checkpoints[i] = $root.repository.Checkpoint.fromObject(object.asset_checkpoints[i]);
                }
            }
            if (object.asset_dependencies) {
                if (!Array.isArray(object.asset_dependencies))
                    throw TypeError(".repository.ProjectData.asset_dependencies: array expected");
                message.asset_dependencies = [];
                for (let i = 0; i < object.asset_dependencies.length; ++i) {
                    if (typeof object.asset_dependencies[i] !== "object")
                        throw TypeError(".repository.ProjectData.asset_dependencies: object expected");
                    message.asset_dependencies[i] = $root.repository.AssetDependency.fromObject(object.asset_dependencies[i]);
                }
            }
            if (object.collection_dependencies) {
                if (!Array.isArray(object.collection_dependencies))
                    throw TypeError(".repository.ProjectData.collection_dependencies: array expected");
                message.collection_dependencies = [];
                for (let i = 0; i < object.collection_dependencies.length; ++i) {
                    if (typeof object.collection_dependencies[i] !== "object")
                        throw TypeError(".repository.ProjectData.collection_dependencies: object expected");
                    message.collection_dependencies[i] = $root.repository.CollectionDependency.fromObject(object.collection_dependencies[i]);
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
            if (object.collection_types) {
                if (!Array.isArray(object.collection_types))
                    throw TypeError(".repository.ProjectData.collection_types: array expected");
                message.collection_types = [];
                for (let i = 0; i < object.collection_types.length; ++i) {
                    if (typeof object.collection_types[i] !== "object")
                        throw TypeError(".repository.ProjectData.collection_types: object expected");
                    message.collection_types[i] = $root.repository.CollectionType.fromObject(object.collection_types[i]);
                }
            }
            if (object.collections) {
                if (!Array.isArray(object.collections))
                    throw TypeError(".repository.ProjectData.collections: array expected");
                message.collections = [];
                for (let i = 0; i < object.collections.length; ++i) {
                    if (typeof object.collections[i] !== "object")
                        throw TypeError(".repository.ProjectData.collections: object expected");
                    message.collections[i] = $root.repository.Collection.fromObject(object.collections[i]);
                }
            }
            if (object.collection_assignees) {
                if (!Array.isArray(object.collection_assignees))
                    throw TypeError(".repository.ProjectData.collection_assignees: array expected");
                message.collection_assignees = [];
                for (let i = 0; i < object.collection_assignees.length; ++i) {
                    if (typeof object.collection_assignees[i] !== "object")
                        throw TypeError(".repository.ProjectData.collection_assignees: object expected");
                    message.collection_assignees[i] = $root.repository.CollectionAssignee.fromObject(object.collection_assignees[i]);
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
            if (object.asset_tags) {
                if (!Array.isArray(object.asset_tags))
                    throw TypeError(".repository.ProjectData.asset_tags: array expected");
                message.asset_tags = [];
                for (let i = 0; i < object.asset_tags.length; ++i) {
                    if (typeof object.asset_tags[i] !== "object")
                        throw TypeError(".repository.ProjectData.asset_tags: object expected");
                    message.asset_tags[i] = $root.repository.AssetTag.fromObject(object.asset_tags[i]);
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
            if (object.workflow_collections) {
                if (!Array.isArray(object.workflow_collections))
                    throw TypeError(".repository.ProjectData.workflow_collections: array expected");
                message.workflow_collections = [];
                for (let i = 0; i < object.workflow_collections.length; ++i) {
                    if (typeof object.workflow_collections[i] !== "object")
                        throw TypeError(".repository.ProjectData.workflow_collections: object expected");
                    message.workflow_collections[i] = $root.repository.WorkflowCollection.fromObject(object.workflow_collections[i]);
                }
            }
            if (object.workflow_assets) {
                if (!Array.isArray(object.workflow_assets))
                    throw TypeError(".repository.ProjectData.workflow_assets: array expected");
                message.workflow_assets = [];
                for (let i = 0; i < object.workflow_assets.length; ++i) {
                    if (typeof object.workflow_assets[i] !== "object")
                        throw TypeError(".repository.ProjectData.workflow_assets: object expected");
                    message.workflow_assets[i] = $root.repository.WorkflowAsset.fromObject(object.workflow_assets[i]);
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
            if (object.integration_projects) {
                if (!Array.isArray(object.integration_projects))
                    throw TypeError(".repository.ProjectData.integration_projects: array expected");
                message.integration_projects = [];
                for (let i = 0; i < object.integration_projects.length; ++i) {
                    if (typeof object.integration_projects[i] !== "object")
                        throw TypeError(".repository.ProjectData.integration_projects: object expected");
                    message.integration_projects[i] = $root.repository.IntegrationProject.fromObject(object.integration_projects[i]);
                }
            }
            if (object.integration_collection_mappings) {
                if (!Array.isArray(object.integration_collection_mappings))
                    throw TypeError(".repository.ProjectData.integration_collection_mappings: array expected");
                message.integration_collection_mappings = [];
                for (let i = 0; i < object.integration_collection_mappings.length; ++i) {
                    if (typeof object.integration_collection_mappings[i] !== "object")
                        throw TypeError(".repository.ProjectData.integration_collection_mappings: object expected");
                    message.integration_collection_mappings[i] = $root.repository.IntegrationCollectionMapping.fromObject(object.integration_collection_mappings[i]);
                }
            }
            if (object.integration_asset_mappings) {
                if (!Array.isArray(object.integration_asset_mappings))
                    throw TypeError(".repository.ProjectData.integration_asset_mappings: array expected");
                message.integration_asset_mappings = [];
                for (let i = 0; i < object.integration_asset_mappings.length; ++i) {
                    if (typeof object.integration_asset_mappings[i] !== "object")
                        throw TypeError(".repository.ProjectData.integration_asset_mappings: object expected");
                    message.integration_asset_mappings[i] = $root.repository.IntegrationAssetMapping.fromObject(object.integration_asset_mappings[i]);
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
                object.assets = [];
                object.asset_types = [];
                object.asset_checkpoints = [];
                object.asset_dependencies = [];
                object.collection_dependencies = [];
                object.statuses = [];
                object.dependency_types = [];
                object.users = [];
                object.roles = [];
                object.collection_types = [];
                object.collections = [];
                object.collection_assignees = [];
                object.templates = [];
                object.tags = [];
                object.asset_tags = [];
                object.workflows = [];
                object.workflow_links = [];
                object.workflow_collections = [];
                object.workflow_assets = [];
                object.tomb = [];
                object.integration_projects = [];
                object.integration_collection_mappings = [];
                object.integration_asset_mappings = [];
            }
            if (options.defaults)
                object.project_preview = "";
            if (message.project_preview != null && message.hasOwnProperty("project_preview"))
                object.project_preview = message.project_preview;
            if (message.assets && message.assets.length) {
                object.assets = [];
                for (let j = 0; j < message.assets.length; ++j)
                    object.assets[j] = $root.repository.Asset.toObject(message.assets[j], options);
            }
            if (message.asset_types && message.asset_types.length) {
                object.asset_types = [];
                for (let j = 0; j < message.asset_types.length; ++j)
                    object.asset_types[j] = $root.repository.AssetType.toObject(message.asset_types[j], options);
            }
            if (message.asset_checkpoints && message.asset_checkpoints.length) {
                object.asset_checkpoints = [];
                for (let j = 0; j < message.asset_checkpoints.length; ++j)
                    object.asset_checkpoints[j] = $root.repository.Checkpoint.toObject(message.asset_checkpoints[j], options);
            }
            if (message.asset_dependencies && message.asset_dependencies.length) {
                object.asset_dependencies = [];
                for (let j = 0; j < message.asset_dependencies.length; ++j)
                    object.asset_dependencies[j] = $root.repository.AssetDependency.toObject(message.asset_dependencies[j], options);
            }
            if (message.collection_dependencies && message.collection_dependencies.length) {
                object.collection_dependencies = [];
                for (let j = 0; j < message.collection_dependencies.length; ++j)
                    object.collection_dependencies[j] = $root.repository.CollectionDependency.toObject(message.collection_dependencies[j], options);
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
            if (message.collection_types && message.collection_types.length) {
                object.collection_types = [];
                for (let j = 0; j < message.collection_types.length; ++j)
                    object.collection_types[j] = $root.repository.CollectionType.toObject(message.collection_types[j], options);
            }
            if (message.collections && message.collections.length) {
                object.collections = [];
                for (let j = 0; j < message.collections.length; ++j)
                    object.collections[j] = $root.repository.Collection.toObject(message.collections[j], options);
            }
            if (message.collection_assignees && message.collection_assignees.length) {
                object.collection_assignees = [];
                for (let j = 0; j < message.collection_assignees.length; ++j)
                    object.collection_assignees[j] = $root.repository.CollectionAssignee.toObject(message.collection_assignees[j], options);
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
            if (message.asset_tags && message.asset_tags.length) {
                object.asset_tags = [];
                for (let j = 0; j < message.asset_tags.length; ++j)
                    object.asset_tags[j] = $root.repository.AssetTag.toObject(message.asset_tags[j], options);
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
            if (message.workflow_collections && message.workflow_collections.length) {
                object.workflow_collections = [];
                for (let j = 0; j < message.workflow_collections.length; ++j)
                    object.workflow_collections[j] = $root.repository.WorkflowCollection.toObject(message.workflow_collections[j], options);
            }
            if (message.workflow_assets && message.workflow_assets.length) {
                object.workflow_assets = [];
                for (let j = 0; j < message.workflow_assets.length; ++j)
                    object.workflow_assets[j] = $root.repository.WorkflowAsset.toObject(message.workflow_assets[j], options);
            }
            if (message.tomb && message.tomb.length) {
                object.tomb = [];
                for (let j = 0; j < message.tomb.length; ++j)
                    object.tomb[j] = $root.repository.Tomb.toObject(message.tomb[j], options);
            }
            if (message.integration_projects && message.integration_projects.length) {
                object.integration_projects = [];
                for (let j = 0; j < message.integration_projects.length; ++j)
                    object.integration_projects[j] = $root.repository.IntegrationProject.toObject(message.integration_projects[j], options);
            }
            if (message.integration_collection_mappings && message.integration_collection_mappings.length) {
                object.integration_collection_mappings = [];
                for (let j = 0; j < message.integration_collection_mappings.length; ++j)
                    object.integration_collection_mappings[j] = $root.repository.IntegrationCollectionMapping.toObject(message.integration_collection_mappings[j], options);
            }
            if (message.integration_asset_mappings && message.integration_asset_mappings.length) {
                object.integration_asset_mappings = [];
                for (let j = 0; j < message.integration_asset_mappings.length; ++j)
                    object.integration_asset_mappings[j] = $root.repository.IntegrationAssetMapping.toObject(message.integration_asset_mappings[j], options);
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

    repository.FullAsset = (function() {

        /**
         * Properties of a FullAsset.
         * @memberof repository
         * @interface IFullAsset
         * @property {string|null} [id] FullAsset id
         * @property {number|Long|null} [mtime] FullAsset mtime
         * @property {string|null} [created_at] FullAsset created_at
         * @property {string|null} [name] FullAsset name
         * @property {string|null} [description] FullAsset description
         * @property {string|null} [extension] FullAsset extension
         * @property {boolean|null} [is_resource] FullAsset is_resource
         * @property {string|null} [status_id] FullAsset status_id
         * @property {string|null} [status_short_name] FullAsset status_short_name
         * @property {string|null} [asset_type_id] FullAsset asset_type_id
         * @property {string|null} [asset_type_name] FullAsset asset_type_name
         * @property {string|null} [asset_type_icon] FullAsset asset_type_icon
         * @property {string|null} [collection_id] FullAsset collection_id
         * @property {string|null} [collection_name] FullAsset collection_name
         * @property {string|null} [collection_path] FullAsset collection_path
         * @property {string|null} [asset_path] FullAsset asset_path
         * @property {string|null} [assignee_id] FullAsset assignee_id
         * @property {string|null} [assignee_email] FullAsset assignee_email
         * @property {string|null} [assignee_name] FullAsset assignee_name
         * @property {string|null} [assigner_id] FullAsset assigner_id
         * @property {string|null} [assigner_email] FullAsset assigner_email
         * @property {string|null} [assigner_name] FullAsset assigner_name
         * @property {boolean|null} [is_dependency] FullAsset is_dependency
         * @property {number|null} [dependency_level] FullAsset dependency_level
         * @property {string|null} [file_path] FullAsset file_path
         * @property {Array.<string>|null} [tags] FullAsset tags
         * @property {string|null} [tags_raw] FullAsset tags_raw
         * @property {Array.<string>|null} [collection_dependencies] FullAsset collection_dependencies
         * @property {string|null} [collection_dependencies_raw] FullAsset collection_dependencies_raw
         * @property {Array.<string>|null} [dependencies] FullAsset dependencies
         * @property {string|null} [dependencies_raw] FullAsset dependencies_raw
         * @property {string|null} [file_status] FullAsset file_status
         * @property {repository.IStatus|null} [status] FullAsset status
         * @property {boolean|null} [is_link] FullAsset is_link
         * @property {string|null} [pointer] FullAsset pointer
         * @property {string|null} [preview_id] FullAsset preview_id
         * @property {Uint8Array|null} [preview] FullAsset preview
         * @property {string|null} [preview_extension] FullAsset preview_extension
         * @property {Array.<repository.ICheckpoint>|null} [checkpoints] FullAsset checkpoints
         * @property {boolean|null} [trashed] FullAsset trashed
         * @property {boolean|null} [synced] FullAsset synced
         * @property {string|null} [type] FullAsset type
         */

        /**
         * Constructs a new FullAsset.
         * @memberof repository
         * @classdesc Represents a FullAsset.
         * @implements IFullAsset
         * @constructor
         * @param {repository.IFullAsset=} [properties] Properties to set
         */
        function FullAsset(properties) {
            this.tags = [];
            this.collection_dependencies = [];
            this.dependencies = [];
            this.checkpoints = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * FullAsset id.
         * @member {string} id
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.id = "";

        /**
         * FullAsset mtime.
         * @member {number|Long} mtime
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.mtime = $util.Long ? $util.Long.fromBits(0,0,false) : 0;

        /**
         * FullAsset created_at.
         * @member {string} created_at
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.created_at = "";

        /**
         * FullAsset name.
         * @member {string} name
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.name = "";

        /**
         * FullAsset description.
         * @member {string} description
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.description = "";

        /**
         * FullAsset extension.
         * @member {string} extension
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.extension = "";

        /**
         * FullAsset is_resource.
         * @member {boolean} is_resource
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.is_resource = false;

        /**
         * FullAsset status_id.
         * @member {string} status_id
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.status_id = "";

        /**
         * FullAsset status_short_name.
         * @member {string} status_short_name
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.status_short_name = "";

        /**
         * FullAsset asset_type_id.
         * @member {string} asset_type_id
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.asset_type_id = "";

        /**
         * FullAsset asset_type_name.
         * @member {string} asset_type_name
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.asset_type_name = "";

        /**
         * FullAsset asset_type_icon.
         * @member {string} asset_type_icon
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.asset_type_icon = "";

        /**
         * FullAsset collection_id.
         * @member {string} collection_id
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.collection_id = "";

        /**
         * FullAsset collection_name.
         * @member {string} collection_name
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.collection_name = "";

        /**
         * FullAsset collection_path.
         * @member {string} collection_path
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.collection_path = "";

        /**
         * FullAsset asset_path.
         * @member {string} asset_path
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.asset_path = "";

        /**
         * FullAsset assignee_id.
         * @member {string} assignee_id
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.assignee_id = "";

        /**
         * FullAsset assignee_email.
         * @member {string} assignee_email
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.assignee_email = "";

        /**
         * FullAsset assignee_name.
         * @member {string} assignee_name
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.assignee_name = "";

        /**
         * FullAsset assigner_id.
         * @member {string} assigner_id
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.assigner_id = "";

        /**
         * FullAsset assigner_email.
         * @member {string} assigner_email
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.assigner_email = "";

        /**
         * FullAsset assigner_name.
         * @member {string} assigner_name
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.assigner_name = "";

        /**
         * FullAsset is_dependency.
         * @member {boolean} is_dependency
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.is_dependency = false;

        /**
         * FullAsset dependency_level.
         * @member {number} dependency_level
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.dependency_level = 0;

        /**
         * FullAsset file_path.
         * @member {string} file_path
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.file_path = "";

        /**
         * FullAsset tags.
         * @member {Array.<string>} tags
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.tags = $util.emptyArray;

        /**
         * FullAsset tags_raw.
         * @member {string} tags_raw
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.tags_raw = "";

        /**
         * FullAsset collection_dependencies.
         * @member {Array.<string>} collection_dependencies
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.collection_dependencies = $util.emptyArray;

        /**
         * FullAsset collection_dependencies_raw.
         * @member {string} collection_dependencies_raw
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.collection_dependencies_raw = "";

        /**
         * FullAsset dependencies.
         * @member {Array.<string>} dependencies
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.dependencies = $util.emptyArray;

        /**
         * FullAsset dependencies_raw.
         * @member {string} dependencies_raw
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.dependencies_raw = "";

        /**
         * FullAsset file_status.
         * @member {string} file_status
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.file_status = "";

        /**
         * FullAsset status.
         * @member {repository.IStatus|null|undefined} status
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.status = null;

        /**
         * FullAsset is_link.
         * @member {boolean} is_link
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.is_link = false;

        /**
         * FullAsset pointer.
         * @member {string} pointer
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.pointer = "";

        /**
         * FullAsset preview_id.
         * @member {string} preview_id
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.preview_id = "";

        /**
         * FullAsset preview.
         * @member {Uint8Array} preview
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.preview = $util.newBuffer([]);

        /**
         * FullAsset preview_extension.
         * @member {string} preview_extension
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.preview_extension = "";

        /**
         * FullAsset checkpoints.
         * @member {Array.<repository.ICheckpoint>} checkpoints
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.checkpoints = $util.emptyArray;

        /**
         * FullAsset trashed.
         * @member {boolean} trashed
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.trashed = false;

        /**
         * FullAsset synced.
         * @member {boolean} synced
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.synced = false;

        /**
         * FullAsset type.
         * @member {string} type
         * @memberof repository.FullAsset
         * @instance
         */
        FullAsset.prototype.type = "";

        /**
         * Creates a new FullAsset instance using the specified properties.
         * @function create
         * @memberof repository.FullAsset
         * @static
         * @param {repository.IFullAsset=} [properties] Properties to set
         * @returns {repository.FullAsset} FullAsset instance
         */
        FullAsset.create = function create(properties) {
            return new FullAsset(properties);
        };

        /**
         * Encodes the specified FullAsset message. Does not implicitly {@link repository.FullAsset.verify|verify} messages.
         * @function encode
         * @memberof repository.FullAsset
         * @static
         * @param {repository.IFullAsset} message FullAsset message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FullAsset.encode = function encode(message, writer) {
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
            if (message.asset_type_id != null && Object.hasOwnProperty.call(message, "asset_type_id"))
                writer.uint32(/* id 10, wireType 2 =*/82).string(message.asset_type_id);
            if (message.asset_type_name != null && Object.hasOwnProperty.call(message, "asset_type_name"))
                writer.uint32(/* id 11, wireType 2 =*/90).string(message.asset_type_name);
            if (message.asset_type_icon != null && Object.hasOwnProperty.call(message, "asset_type_icon"))
                writer.uint32(/* id 12, wireType 2 =*/98).string(message.asset_type_icon);
            if (message.collection_id != null && Object.hasOwnProperty.call(message, "collection_id"))
                writer.uint32(/* id 13, wireType 2 =*/106).string(message.collection_id);
            if (message.collection_name != null && Object.hasOwnProperty.call(message, "collection_name"))
                writer.uint32(/* id 14, wireType 2 =*/114).string(message.collection_name);
            if (message.collection_path != null && Object.hasOwnProperty.call(message, "collection_path"))
                writer.uint32(/* id 15, wireType 2 =*/122).string(message.collection_path);
            if (message.asset_path != null && Object.hasOwnProperty.call(message, "asset_path"))
                writer.uint32(/* id 16, wireType 2 =*/130).string(message.asset_path);
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
            if (message.collection_dependencies != null && message.collection_dependencies.length)
                for (let i = 0; i < message.collection_dependencies.length; ++i)
                    writer.uint32(/* id 28, wireType 2 =*/226).string(message.collection_dependencies[i]);
            if (message.collection_dependencies_raw != null && Object.hasOwnProperty.call(message, "collection_dependencies_raw"))
                writer.uint32(/* id 29, wireType 2 =*/234).string(message.collection_dependencies_raw);
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
         * Encodes the specified FullAsset message, length delimited. Does not implicitly {@link repository.FullAsset.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.FullAsset
         * @static
         * @param {repository.IFullAsset} message FullAsset message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FullAsset.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a FullAsset message from the specified reader or buffer.
         * @function decode
         * @memberof repository.FullAsset
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.FullAsset} FullAsset
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FullAsset.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.FullAsset();
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
                        message.asset_type_id = reader.string();
                        break;
                    }
                case 11: {
                        message.asset_type_name = reader.string();
                        break;
                    }
                case 12: {
                        message.asset_type_icon = reader.string();
                        break;
                    }
                case 13: {
                        message.collection_id = reader.string();
                        break;
                    }
                case 14: {
                        message.collection_name = reader.string();
                        break;
                    }
                case 15: {
                        message.collection_path = reader.string();
                        break;
                    }
                case 16: {
                        message.asset_path = reader.string();
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
                        if (!(message.collection_dependencies && message.collection_dependencies.length))
                            message.collection_dependencies = [];
                        message.collection_dependencies.push(reader.string());
                        break;
                    }
                case 29: {
                        message.collection_dependencies_raw = reader.string();
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
         * Decodes a FullAsset message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.FullAsset
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.FullAsset} FullAsset
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FullAsset.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a FullAsset message.
         * @function verify
         * @memberof repository.FullAsset
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        FullAsset.verify = function verify(message) {
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
            if (message.asset_type_id != null && message.hasOwnProperty("asset_type_id"))
                if (!$util.isString(message.asset_type_id))
                    return "asset_type_id: string expected";
            if (message.asset_type_name != null && message.hasOwnProperty("asset_type_name"))
                if (!$util.isString(message.asset_type_name))
                    return "asset_type_name: string expected";
            if (message.asset_type_icon != null && message.hasOwnProperty("asset_type_icon"))
                if (!$util.isString(message.asset_type_icon))
                    return "asset_type_icon: string expected";
            if (message.collection_id != null && message.hasOwnProperty("collection_id"))
                if (!$util.isString(message.collection_id))
                    return "collection_id: string expected";
            if (message.collection_name != null && message.hasOwnProperty("collection_name"))
                if (!$util.isString(message.collection_name))
                    return "collection_name: string expected";
            if (message.collection_path != null && message.hasOwnProperty("collection_path"))
                if (!$util.isString(message.collection_path))
                    return "collection_path: string expected";
            if (message.asset_path != null && message.hasOwnProperty("asset_path"))
                if (!$util.isString(message.asset_path))
                    return "asset_path: string expected";
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
            if (message.collection_dependencies != null && message.hasOwnProperty("collection_dependencies")) {
                if (!Array.isArray(message.collection_dependencies))
                    return "collection_dependencies: array expected";
                for (let i = 0; i < message.collection_dependencies.length; ++i)
                    if (!$util.isString(message.collection_dependencies[i]))
                        return "collection_dependencies: string[] expected";
            }
            if (message.collection_dependencies_raw != null && message.hasOwnProperty("collection_dependencies_raw"))
                if (!$util.isString(message.collection_dependencies_raw))
                    return "collection_dependencies_raw: string expected";
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
         * Creates a FullAsset message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.FullAsset
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.FullAsset} FullAsset
         */
        FullAsset.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.FullAsset)
                return object;
            let message = new $root.repository.FullAsset();
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
            if (object.asset_type_id != null)
                message.asset_type_id = String(object.asset_type_id);
            if (object.asset_type_name != null)
                message.asset_type_name = String(object.asset_type_name);
            if (object.asset_type_icon != null)
                message.asset_type_icon = String(object.asset_type_icon);
            if (object.collection_id != null)
                message.collection_id = String(object.collection_id);
            if (object.collection_name != null)
                message.collection_name = String(object.collection_name);
            if (object.collection_path != null)
                message.collection_path = String(object.collection_path);
            if (object.asset_path != null)
                message.asset_path = String(object.asset_path);
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
                    throw TypeError(".repository.FullAsset.tags: array expected");
                message.tags = [];
                for (let i = 0; i < object.tags.length; ++i)
                    message.tags[i] = String(object.tags[i]);
            }
            if (object.tags_raw != null)
                message.tags_raw = String(object.tags_raw);
            if (object.collection_dependencies) {
                if (!Array.isArray(object.collection_dependencies))
                    throw TypeError(".repository.FullAsset.collection_dependencies: array expected");
                message.collection_dependencies = [];
                for (let i = 0; i < object.collection_dependencies.length; ++i)
                    message.collection_dependencies[i] = String(object.collection_dependencies[i]);
            }
            if (object.collection_dependencies_raw != null)
                message.collection_dependencies_raw = String(object.collection_dependencies_raw);
            if (object.dependencies) {
                if (!Array.isArray(object.dependencies))
                    throw TypeError(".repository.FullAsset.dependencies: array expected");
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
                    throw TypeError(".repository.FullAsset.status: object expected");
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
                    throw TypeError(".repository.FullAsset.checkpoints: array expected");
                message.checkpoints = [];
                for (let i = 0; i < object.checkpoints.length; ++i) {
                    if (typeof object.checkpoints[i] !== "object")
                        throw TypeError(".repository.FullAsset.checkpoints: object expected");
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
         * Creates a plain object from a FullAsset message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.FullAsset
         * @static
         * @param {repository.FullAsset} message FullAsset
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        FullAsset.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults) {
                object.tags = [];
                object.collection_dependencies = [];
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
                object.asset_type_id = "";
                object.asset_type_name = "";
                object.asset_type_icon = "";
                object.collection_id = "";
                object.collection_name = "";
                object.collection_path = "";
                object.asset_path = "";
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
                object.collection_dependencies_raw = "";
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
            if (message.asset_type_id != null && message.hasOwnProperty("asset_type_id"))
                object.asset_type_id = message.asset_type_id;
            if (message.asset_type_name != null && message.hasOwnProperty("asset_type_name"))
                object.asset_type_name = message.asset_type_name;
            if (message.asset_type_icon != null && message.hasOwnProperty("asset_type_icon"))
                object.asset_type_icon = message.asset_type_icon;
            if (message.collection_id != null && message.hasOwnProperty("collection_id"))
                object.collection_id = message.collection_id;
            if (message.collection_name != null && message.hasOwnProperty("collection_name"))
                object.collection_name = message.collection_name;
            if (message.collection_path != null && message.hasOwnProperty("collection_path"))
                object.collection_path = message.collection_path;
            if (message.asset_path != null && message.hasOwnProperty("asset_path"))
                object.asset_path = message.asset_path;
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
            if (message.collection_dependencies && message.collection_dependencies.length) {
                object.collection_dependencies = [];
                for (let j = 0; j < message.collection_dependencies.length; ++j)
                    object.collection_dependencies[j] = message.collection_dependencies[j];
            }
            if (message.collection_dependencies_raw != null && message.hasOwnProperty("collection_dependencies_raw"))
                object.collection_dependencies_raw = message.collection_dependencies_raw;
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
         * Converts this FullAsset to JSON.
         * @function toJSON
         * @memberof repository.FullAsset
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        FullAsset.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for FullAsset
         * @function getTypeUrl
         * @memberof repository.FullAsset
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        FullAsset.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.FullAsset";
        };

        return FullAsset;
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

    repository.FullAssetList = (function() {

        /**
         * Properties of a FullAssetList.
         * @memberof repository
         * @interface IFullAssetList
         * @property {Array.<repository.IFullAsset>|null} [full_assets] FullAssetList full_assets
         */

        /**
         * Constructs a new FullAssetList.
         * @memberof repository
         * @classdesc Represents a FullAssetList.
         * @implements IFullAssetList
         * @constructor
         * @param {repository.IFullAssetList=} [properties] Properties to set
         */
        function FullAssetList(properties) {
            this.full_assets = [];
            if (properties)
                for (let keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * FullAssetList full_assets.
         * @member {Array.<repository.IFullAsset>} full_assets
         * @memberof repository.FullAssetList
         * @instance
         */
        FullAssetList.prototype.full_assets = $util.emptyArray;

        /**
         * Creates a new FullAssetList instance using the specified properties.
         * @function create
         * @memberof repository.FullAssetList
         * @static
         * @param {repository.IFullAssetList=} [properties] Properties to set
         * @returns {repository.FullAssetList} FullAssetList instance
         */
        FullAssetList.create = function create(properties) {
            return new FullAssetList(properties);
        };

        /**
         * Encodes the specified FullAssetList message. Does not implicitly {@link repository.FullAssetList.verify|verify} messages.
         * @function encode
         * @memberof repository.FullAssetList
         * @static
         * @param {repository.IFullAssetList} message FullAssetList message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FullAssetList.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.full_assets != null && message.full_assets.length)
                for (let i = 0; i < message.full_assets.length; ++i)
                    $root.repository.FullAsset.encode(message.full_assets[i], writer.uint32(/* id 1, wireType 2 =*/10).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified FullAssetList message, length delimited. Does not implicitly {@link repository.FullAssetList.verify|verify} messages.
         * @function encodeDelimited
         * @memberof repository.FullAssetList
         * @static
         * @param {repository.IFullAssetList} message FullAssetList message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        FullAssetList.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes a FullAssetList message from the specified reader or buffer.
         * @function decode
         * @memberof repository.FullAssetList
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {repository.FullAssetList} FullAssetList
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FullAssetList.decode = function decode(reader, length, error) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            let end = length === undefined ? reader.len : reader.pos + length, message = new $root.repository.FullAssetList();
            while (reader.pos < end) {
                let tag = reader.uint32();
                if (tag === error)
                    break;
                switch (tag >>> 3) {
                case 1: {
                        if (!(message.full_assets && message.full_assets.length))
                            message.full_assets = [];
                        message.full_assets.push($root.repository.FullAsset.decode(reader, reader.uint32()));
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
         * Decodes a FullAssetList message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof repository.FullAssetList
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {repository.FullAssetList} FullAssetList
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        FullAssetList.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies a FullAssetList message.
         * @function verify
         * @memberof repository.FullAssetList
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        FullAssetList.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.full_assets != null && message.hasOwnProperty("full_assets")) {
                if (!Array.isArray(message.full_assets))
                    return "full_assets: array expected";
                for (let i = 0; i < message.full_assets.length; ++i) {
                    let error = $root.repository.FullAsset.verify(message.full_assets[i]);
                    if (error)
                        return "full_assets." + error;
                }
            }
            return null;
        };

        /**
         * Creates a FullAssetList message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof repository.FullAssetList
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {repository.FullAssetList} FullAssetList
         */
        FullAssetList.fromObject = function fromObject(object) {
            if (object instanceof $root.repository.FullAssetList)
                return object;
            let message = new $root.repository.FullAssetList();
            if (object.full_assets) {
                if (!Array.isArray(object.full_assets))
                    throw TypeError(".repository.FullAssetList.full_assets: array expected");
                message.full_assets = [];
                for (let i = 0; i < object.full_assets.length; ++i) {
                    if (typeof object.full_assets[i] !== "object")
                        throw TypeError(".repository.FullAssetList.full_assets: object expected");
                    message.full_assets[i] = $root.repository.FullAsset.fromObject(object.full_assets[i]);
                }
            }
            return message;
        };

        /**
         * Creates a plain object from a FullAssetList message. Also converts values to other types if specified.
         * @function toObject
         * @memberof repository.FullAssetList
         * @static
         * @param {repository.FullAssetList} message FullAssetList
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        FullAssetList.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            let object = {};
            if (options.arrays || options.defaults)
                object.full_assets = [];
            if (message.full_assets && message.full_assets.length) {
                object.full_assets = [];
                for (let j = 0; j < message.full_assets.length; ++j)
                    object.full_assets[j] = $root.repository.FullAsset.toObject(message.full_assets[j], options);
            }
            return object;
        };

        /**
         * Converts this FullAssetList to JSON.
         * @function toJSON
         * @memberof repository.FullAssetList
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        FullAssetList.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        /**
         * Gets the default type url for FullAssetList
         * @function getTypeUrl
         * @memberof repository.FullAssetList
         * @static
         * @param {string} [typeUrlPrefix] your custom typeUrlPrefix(default "type.googleapis.com")
         * @returns {string} The default type url
         */
        FullAssetList.getTypeUrl = function getTypeUrl(typeUrlPrefix) {
            if (typeUrlPrefix === undefined) {
                typeUrlPrefix = "type.googleapis.com";
            }
            return typeUrlPrefix + "/repository.FullAssetList";
        };

        return FullAssetList;
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
