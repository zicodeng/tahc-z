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
            let channel1 = new Channel('General', 'test general channel', 'Zico Deng');
            let channel2 = new Channel('Hello', 'test hello channel', 'Zico Deng');

            return store
                .insert(channel1)
                .then(newChannel => {
                    expect(newChannel._id).toBeDefined();
                    return store.insert(channel2);
                })
                .then(newChannel => {
                    expect(newChannel._id).toBeDefined();
                })
                .then(() => {
                    return store.getAll();
                })
                .then(channels => {
                    expect(channels.length).toEqual(2);
                    expect(channels[0]).toEqual(channel1);
                    expect(channels[1]).toEqual(channel2);
                    return channel1._id;
                })
                .then(channel1ID => {
                    return store.get(channel1ID);
                })
                .then(fetchedChannel => {
                    expect(fetchedChannel).toEqual(channel1);
                    const updates = {
                        name: 'Info 344'
                    };
                    return store.update(channel1._id, updates);
                })
                .then(updatedChannel => {
                    expect(updatedChannel.name).toBe('Info 344');
                    return store.delete(channel2._id);
                })
                .then(deletedChannel => {
                    return store.get(deletedChannel._id);
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
