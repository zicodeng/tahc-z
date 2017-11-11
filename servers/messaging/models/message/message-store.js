'use strict';

const mongodb = require('mongodb');

class MessageStore {
    constructor(db, colName) {
        this.collection = db.collection(colName);
    }

    // insert() creates a new message object in MongoDB.
    insert(message) {
        message._id = new mongodb.ObjectID();
        return this.collection.insertOne(message).then(() => message);
    }

    // get() retrieves one message object from MongoDB for a given message ID.
    get(id) {
        return this.collection.findOne({ _id: id });
    }

    // getAll() retrieves 50 message objects from MongoDB for a given channel ID.
    getAll(channelID) {
        return this.collection
            .find({ channelID: channelID })
            .limit(50)
            .toArray();
    }

    // update() updates a message object for a given message ID.
    // It returns the updated message object.
    update(id, updates) {
        let updateDoc = {
            $set: updates
        };
        return this.collection
            .findOneAndUpdate({ _id: id }, updateDoc, { returnOriginal: false })
            .then(result => {
                return result.value;
            });
    }

    // delete() deletes a message object for a given message ID.
    delete(id) {
        return this.collection.deleteOne({ _id: id });
    }
}

module.exports = MessageStore;
