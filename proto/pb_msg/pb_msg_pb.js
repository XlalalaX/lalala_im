// source: pb_msg.proto
/**
 * @fileoverview
 * @enhanceable
 * @suppress {messageConventions} JS Compiler reports an error if a variable or
 *     field starts with 'MSG_' and isn't a translatable message.
 * @public
 */
// GENERATED CODE -- DO NOT EDIT!

var jspb = require('google-protobuf');
var goog = jspb;
var global = Function('return this')();

goog.exportSymbol('proto.pb_msg.ErrRsp', null, global);
goog.exportSymbol('proto.pb_msg.Msg', null, global);
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.pb_msg.Msg = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pb_msg.Msg, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pb_msg.Msg.displayName = 'proto.pb_msg.Msg';
}
/**
 * Generated by JsPbCodeGenerator.
 * @param {Array=} opt_data Optional initial data array, typically from a
 * server response, or constructed directly in Javascript. The array is used
 * in place and becomes part of the constructed object. It is not cloned.
 * If no data is provided, the constructed object will be empty, but still
 * valid.
 * @extends {jspb.Message}
 * @constructor
 */
proto.pb_msg.ErrRsp = function(opt_data) {
  jspb.Message.initialize(this, opt_data, 0, -1, null, null);
};
goog.inherits(proto.pb_msg.ErrRsp, jspb.Message);
if (goog.DEBUG && !COMPILED) {
  /**
   * @public
   * @override
   */
  proto.pb_msg.ErrRsp.displayName = 'proto.pb_msg.ErrRsp';
}



if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.pb_msg.Msg.prototype.toObject = function(opt_includeInstance) {
  return proto.pb_msg.Msg.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pb_msg.Msg} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pb_msg.Msg.toObject = function(includeInstance, msg) {
  var f, obj = {
    sendid: jspb.Message.getFieldWithDefault(msg, 1, ""),
    recvid: jspb.Message.getFieldWithDefault(msg, 2, ""),
    groupid: jspb.Message.getFieldWithDefault(msg, 3, ""),
    senderplatformid: jspb.Message.getFieldWithDefault(msg, 4, 0),
    sendernickname: jspb.Message.getFieldWithDefault(msg, 5, ""),
    senderfaceurl: jspb.Message.getFieldWithDefault(msg, 6, ""),
    sessiontype: jspb.Message.getFieldWithDefault(msg, 7, 0),
    msgfrom: jspb.Message.getFieldWithDefault(msg, 8, 0),
    contenttype: jspb.Message.getFieldWithDefault(msg, 9, 0),
    content: jspb.Message.getFieldWithDefault(msg, 10, ""),
    seq: jspb.Message.getFieldWithDefault(msg, 11, 0),
    sendtime: jspb.Message.getFieldWithDefault(msg, 12, 0),
    status: jspb.Message.getFieldWithDefault(msg, 13, 0),
    file: msg.getFile_asB64()
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.pb_msg.Msg}
 */
proto.pb_msg.Msg.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pb_msg.Msg;
  return proto.pb_msg.Msg.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pb_msg.Msg} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pb_msg.Msg}
 */
proto.pb_msg.Msg.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {string} */ (reader.readString());
      msg.setSendid(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setRecvid(value);
      break;
    case 3:
      var value = /** @type {string} */ (reader.readString());
      msg.setGroupid(value);
      break;
    case 4:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setSenderplatformid(value);
      break;
    case 5:
      var value = /** @type {string} */ (reader.readString());
      msg.setSendernickname(value);
      break;
    case 6:
      var value = /** @type {string} */ (reader.readString());
      msg.setSenderfaceurl(value);
      break;
    case 7:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setSessiontype(value);
      break;
    case 8:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setMsgfrom(value);
      break;
    case 9:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setContenttype(value);
      break;
    case 10:
      var value = /** @type {string} */ (reader.readString());
      msg.setContent(value);
      break;
    case 11:
      var value = /** @type {number} */ (reader.readUint32());
      msg.setSeq(value);
      break;
    case 12:
      var value = /** @type {number} */ (reader.readInt64());
      msg.setSendtime(value);
      break;
    case 13:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setStatus(value);
      break;
    case 14:
      var value = /** @type {!Uint8Array} */ (reader.readBytes());
      msg.setFile(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.pb_msg.Msg.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pb_msg.Msg.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pb_msg.Msg} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pb_msg.Msg.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getSendid();
  if (f.length > 0) {
    writer.writeString(
      1,
      f
    );
  }
  f = message.getRecvid();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
  f = message.getGroupid();
  if (f.length > 0) {
    writer.writeString(
      3,
      f
    );
  }
  f = message.getSenderplatformid();
  if (f !== 0) {
    writer.writeInt32(
      4,
      f
    );
  }
  f = message.getSendernickname();
  if (f.length > 0) {
    writer.writeString(
      5,
      f
    );
  }
  f = message.getSenderfaceurl();
  if (f.length > 0) {
    writer.writeString(
      6,
      f
    );
  }
  f = message.getSessiontype();
  if (f !== 0) {
    writer.writeInt32(
      7,
      f
    );
  }
  f = message.getMsgfrom();
  if (f !== 0) {
    writer.writeInt32(
      8,
      f
    );
  }
  f = message.getContenttype();
  if (f !== 0) {
    writer.writeInt32(
      9,
      f
    );
  }
  f = message.getContent();
  if (f.length > 0) {
    writer.writeString(
      10,
      f
    );
  }
  f = message.getSeq();
  if (f !== 0) {
    writer.writeUint32(
      11,
      f
    );
  }
  f = message.getSendtime();
  if (f !== 0) {
    writer.writeInt64(
      12,
      f
    );
  }
  f = message.getStatus();
  if (f !== 0) {
    writer.writeInt32(
      13,
      f
    );
  }
  f = message.getFile_asU8();
  if (f.length > 0) {
    writer.writeBytes(
      14,
      f
    );
  }
};


/**
 * optional string sendID = 1;
 * @return {string}
 */
proto.pb_msg.Msg.prototype.getSendid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 1, ""));
};


/**
 * @param {string} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setSendid = function(value) {
  return jspb.Message.setProto3StringField(this, 1, value);
};


/**
 * optional string recvID = 2;
 * @return {string}
 */
proto.pb_msg.Msg.prototype.getRecvid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setRecvid = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


/**
 * optional string groupID = 3;
 * @return {string}
 */
proto.pb_msg.Msg.prototype.getGroupid = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 3, ""));
};


/**
 * @param {string} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setGroupid = function(value) {
  return jspb.Message.setProto3StringField(this, 3, value);
};


/**
 * optional int32 senderPlatformID = 4;
 * @return {number}
 */
proto.pb_msg.Msg.prototype.getSenderplatformid = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 4, 0));
};


/**
 * @param {number} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setSenderplatformid = function(value) {
  return jspb.Message.setProto3IntField(this, 4, value);
};


/**
 * optional string senderNickname = 5;
 * @return {string}
 */
proto.pb_msg.Msg.prototype.getSendernickname = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 5, ""));
};


/**
 * @param {string} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setSendernickname = function(value) {
  return jspb.Message.setProto3StringField(this, 5, value);
};


/**
 * optional string senderFaceURL = 6;
 * @return {string}
 */
proto.pb_msg.Msg.prototype.getSenderfaceurl = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 6, ""));
};


/**
 * @param {string} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setSenderfaceurl = function(value) {
  return jspb.Message.setProto3StringField(this, 6, value);
};


/**
 * optional int32 sessionType = 7;
 * @return {number}
 */
proto.pb_msg.Msg.prototype.getSessiontype = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 7, 0));
};


/**
 * @param {number} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setSessiontype = function(value) {
  return jspb.Message.setProto3IntField(this, 7, value);
};


/**
 * optional int32 msgFrom = 8;
 * @return {number}
 */
proto.pb_msg.Msg.prototype.getMsgfrom = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 8, 0));
};


/**
 * @param {number} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setMsgfrom = function(value) {
  return jspb.Message.setProto3IntField(this, 8, value);
};


/**
 * optional int32 contentType = 9;
 * @return {number}
 */
proto.pb_msg.Msg.prototype.getContenttype = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 9, 0));
};


/**
 * @param {number} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setContenttype = function(value) {
  return jspb.Message.setProto3IntField(this, 9, value);
};


/**
 * optional string content = 10;
 * @return {string}
 */
proto.pb_msg.Msg.prototype.getContent = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 10, ""));
};


/**
 * @param {string} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setContent = function(value) {
  return jspb.Message.setProto3StringField(this, 10, value);
};


/**
 * optional uint32 seq = 11;
 * @return {number}
 */
proto.pb_msg.Msg.prototype.getSeq = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 11, 0));
};


/**
 * @param {number} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setSeq = function(value) {
  return jspb.Message.setProto3IntField(this, 11, value);
};


/**
 * optional int64 sendTime = 12;
 * @return {number}
 */
proto.pb_msg.Msg.prototype.getSendtime = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 12, 0));
};


/**
 * @param {number} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setSendtime = function(value) {
  return jspb.Message.setProto3IntField(this, 12, value);
};


/**
 * optional int32 status = 13;
 * @return {number}
 */
proto.pb_msg.Msg.prototype.getStatus = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 13, 0));
};


/**
 * @param {number} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setStatus = function(value) {
  return jspb.Message.setProto3IntField(this, 13, value);
};


/**
 * optional bytes file = 14;
 * @return {!(string|Uint8Array)}
 */
proto.pb_msg.Msg.prototype.getFile = function() {
  return /** @type {!(string|Uint8Array)} */ (jspb.Message.getFieldWithDefault(this, 14, ""));
};


/**
 * optional bytes file = 14;
 * This is a type-conversion wrapper around `getFile()`
 * @return {string}
 */
proto.pb_msg.Msg.prototype.getFile_asB64 = function() {
  return /** @type {string} */ (jspb.Message.bytesAsB64(
      this.getFile()));
};


/**
 * optional bytes file = 14;
 * Note that Uint8Array is not supported on all browsers.
 * @see http://caniuse.com/Uint8Array
 * This is a type-conversion wrapper around `getFile()`
 * @return {!Uint8Array}
 */
proto.pb_msg.Msg.prototype.getFile_asU8 = function() {
  return /** @type {!Uint8Array} */ (jspb.Message.bytesAsU8(
      this.getFile()));
};


/**
 * @param {!(string|Uint8Array)} value
 * @return {!proto.pb_msg.Msg} returns this
 */
proto.pb_msg.Msg.prototype.setFile = function(value) {
  return jspb.Message.setProto3BytesField(this, 14, value);
};





if (jspb.Message.GENERATE_TO_OBJECT) {
/**
 * Creates an object representation of this proto.
 * Field names that are reserved in JavaScript and will be renamed to pb_name.
 * Optional fields that are not set will be set to undefined.
 * To access a reserved field use, foo.pb_<name>, eg, foo.pb_default.
 * For the list of reserved names please see:
 *     net/proto2/compiler/js/internal/generator.cc#kKeyword.
 * @param {boolean=} opt_includeInstance Deprecated. whether to include the
 *     JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @return {!Object}
 */
proto.pb_msg.ErrRsp.prototype.toObject = function(opt_includeInstance) {
  return proto.pb_msg.ErrRsp.toObject(opt_includeInstance, this);
};


/**
 * Static version of the {@see toObject} method.
 * @param {boolean|undefined} includeInstance Deprecated. Whether to include
 *     the JSPB instance for transitional soy proto support:
 *     http://goto/soy-param-migration
 * @param {!proto.pb_msg.ErrRsp} msg The msg instance to transform.
 * @return {!Object}
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pb_msg.ErrRsp.toObject = function(includeInstance, msg) {
  var f, obj = {
    errcode: jspb.Message.getFieldWithDefault(msg, 1, 0),
    errmsg: jspb.Message.getFieldWithDefault(msg, 2, "")
  };

  if (includeInstance) {
    obj.$jspbMessageInstance = msg;
  }
  return obj;
};
}


/**
 * Deserializes binary data (in protobuf wire format).
 * @param {jspb.ByteSource} bytes The bytes to deserialize.
 * @return {!proto.pb_msg.ErrRsp}
 */
proto.pb_msg.ErrRsp.deserializeBinary = function(bytes) {
  var reader = new jspb.BinaryReader(bytes);
  var msg = new proto.pb_msg.ErrRsp;
  return proto.pb_msg.ErrRsp.deserializeBinaryFromReader(msg, reader);
};


/**
 * Deserializes binary data (in protobuf wire format) from the
 * given reader into the given message object.
 * @param {!proto.pb_msg.ErrRsp} msg The message object to deserialize into.
 * @param {!jspb.BinaryReader} reader The BinaryReader to use.
 * @return {!proto.pb_msg.ErrRsp}
 */
proto.pb_msg.ErrRsp.deserializeBinaryFromReader = function(msg, reader) {
  while (reader.nextField()) {
    if (reader.isEndGroup()) {
      break;
    }
    var field = reader.getFieldNumber();
    switch (field) {
    case 1:
      var value = /** @type {number} */ (reader.readInt32());
      msg.setErrcode(value);
      break;
    case 2:
      var value = /** @type {string} */ (reader.readString());
      msg.setErrmsg(value);
      break;
    default:
      reader.skipField();
      break;
    }
  }
  return msg;
};


/**
 * Serializes the message to binary data (in protobuf wire format).
 * @return {!Uint8Array}
 */
proto.pb_msg.ErrRsp.prototype.serializeBinary = function() {
  var writer = new jspb.BinaryWriter();
  proto.pb_msg.ErrRsp.serializeBinaryToWriter(this, writer);
  return writer.getResultBuffer();
};


/**
 * Serializes the given message to binary data (in protobuf wire
 * format), writing to the given BinaryWriter.
 * @param {!proto.pb_msg.ErrRsp} message
 * @param {!jspb.BinaryWriter} writer
 * @suppress {unusedLocalVariables} f is only used for nested messages
 */
proto.pb_msg.ErrRsp.serializeBinaryToWriter = function(message, writer) {
  var f = undefined;
  f = message.getErrcode();
  if (f !== 0) {
    writer.writeInt32(
      1,
      f
    );
  }
  f = message.getErrmsg();
  if (f.length > 0) {
    writer.writeString(
      2,
      f
    );
  }
};


/**
 * optional int32 errCode = 1;
 * @return {number}
 */
proto.pb_msg.ErrRsp.prototype.getErrcode = function() {
  return /** @type {number} */ (jspb.Message.getFieldWithDefault(this, 1, 0));
};


/**
 * @param {number} value
 * @return {!proto.pb_msg.ErrRsp} returns this
 */
proto.pb_msg.ErrRsp.prototype.setErrcode = function(value) {
  return jspb.Message.setProto3IntField(this, 1, value);
};


/**
 * optional string errMsg = 2;
 * @return {string}
 */
proto.pb_msg.ErrRsp.prototype.getErrmsg = function() {
  return /** @type {string} */ (jspb.Message.getFieldWithDefault(this, 2, ""));
};


/**
 * @param {string} value
 * @return {!proto.pb_msg.ErrRsp} returns this
 */
proto.pb_msg.ErrRsp.prototype.setErrmsg = function(value) {
  return jspb.Message.setProto3StringField(this, 2, value);
};


goog.object.extend(exports, proto.pb_msg);
