'use strict';

class Message {
    // Parameter explanation:
    // channelID: ObjectId of channel to which this message belongs.
    // body: the body of the message (text).
    // createdAt: date/time the message was created.
    // creator: copy of the entire profile of the user who created this message.
    // editedAt: date/time the message body was last edited.
    // Note: message object has another property _id, which will be created
    // when we insert it to MongoDB.
    constructor(channelID, body, createdAt, creator, editedAt) {
        this.channelID = channelID;
        this.body = body;
        this.createdAt = createdAt;
        this.creator = editedAt;
    }
}

module.exports = Message;
