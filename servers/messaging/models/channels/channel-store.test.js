'use strict';

const mongodb = require('mongodb');
const mongoAddr = process.env.DBADDR || '192.168.99.100:27017';
const mongoURL = `mongodb://${mongoAddr}/info_344`;

const ChannelStore = require('./channel-store');
const Channel = require('./channel');

describe('Mongo Channel Store', () => {
    test('CRUD Cycle', () => {
        return mongodb.MongoClient.connect(mongoURL).then(db => {
            let store = new ChannelStore(db, 'channels');

            // Create a new Channel.
            let channel = new Channel('General', 'test general channel', 'Zico Deng');

            return store
                .insert(channel)
                .then(channel => {
                    expect(channel._id).toBeDefined();
                    return channel._id;
                })
                .then(channelID => {
                    return store.get(channelID);
                })
                .then(fetchedChannel => {
                    expect(fetchedChannel).toEqual(channel);
                    const updates = {
                        name: 'Info 344'
                    };
                    return store.update(channel._id, updates);
                })
                .then(updatedChannel => {
                    expect(updatedChannel.name).toBe('Info 344');
                    return store.delete(channel._id);
                })
                .then(() => {
                    return store.get(channel._id);
                })
                .then(fetchedChannel => {
                    expect(fetchedChannel).toBeFalsy();
                })
                .then(() => {
                    // If everything went well and all tests passed,
                    // close database connection.
                    db.close();
                })
                .catch(err => {
                    db.close();
                    throw err;
                });
        });
    });
});
