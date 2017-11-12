'use strict';

const mongodb = require('mongodb');
const mongoAddr = process.env.DBADDR || '192.168.99.100:27017';
const mongoURL = `mongodb://${mongoAddr}/info_344`;

const MessageStore = require('./message-store');
const Message = require('./message');

describe('Mongo Message Store', () => {
    test('CRUD Cycle', () => {
        return mongodb.MongoClient.connect(mongoURL).then(db => {
            let store = new MessageStore(db, 'messages');

            const channelID = new mongodb.ObjectID('5a07634f29a1b43820d21e64');
            // Create new message objects.
            const message1 = new Message(channelID, 'Hello!', 'Zico Deng');
            const message2 = new Message(channelID, 'Hola!', 'Zico Deng');

            return store
                .insert(message1)
                .then(newMessage => {
                    expect(newMessage._id).toBeDefined();
                    return store.insert(message2);
                })
                .then(newMessage => {
                    expect(newMessage._id).toBeDefined();
                })
                .then(() => {
                    return store.getAll(channelID);
                })
                .then(messages => {
                    expect(messages.length).toEqual(2);
                    expect(messages[0]).toEqual(message1);
                    expect(messages[1]).toEqual(message2);
                    return message1._id;
                })
                .then(message1ID => {
                    return store.get(message1ID);
                })
                .then(fetchedMessage => {
                    expect(fetchedMessage).toEqual(message1);
                    const updates = {
                        body: 'How are you?'
                    };
                    // Update message1.
                    return store.update(message1._id, updates);
                })
                .then(updatedMessage => {
                    expect(updatedMessage.body).toBe('How are you?');
                    // Delete message2.
                    return store.delete(message2._id);
                })
                .then(deletedMessage => {
                    return store.get(deletedMessage._id);
                })
                .then(fetchedMessage => {
                    expect(fetchedMessage).toBeFalsy();
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
