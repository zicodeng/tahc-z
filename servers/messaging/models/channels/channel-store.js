'use strict';

const mongodb = require('mongodb');

class ChannelStore {
    constructor(db, colName) {
        this.collection = db.collection(colName);
    }

    // insert() creates a new channel object in MongoDB.
    insert(channel) {
        channel._id = new mongodb.ObjectID();
        return this.collection.insertOne(channel).then(() => channel);
    }

    // get() retrieves one channel object from MongoDB for a given channel ID.
    get(id) {
        return this.collection.findOne({ _id: id });
    }

    // getAll() retrieves all channel objects from MongoDB.
    getAll() {
        return this.collection
            .find({})
            .limit(100)
            .toArray();
    }

    // update() updates a channel object for a given channel ID.
    // It returns the updated channel object.
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

    // delete() deletes a channel object for a given channel ID.
    delete(id) {
        return this.collection.deleteOne({ _id: id });
    }
}

module.exports = ChannelStore;
