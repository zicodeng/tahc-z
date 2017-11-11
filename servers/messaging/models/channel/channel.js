'use strict';

class Channel {
    // Parameter explanation:
    // name: unique channel name.
    // description: a short description of the channel.
    // createdAt: date/time the channel was created.
    // creator: copy of the entire profile of the user who created this channel.
    // editedAt: date/time the channel's properties were last edited.
    // Note: channel object has another property _id, which will be created
    // when we insert it to MongoDB.
    constructor(name, description, createdAt, creator, editedAt) {
        this.name = name;
        this.description = description;
        this.createdAt = createdAt;
        this.creator = editedAt;
    }
}

module.exports = Channel;
