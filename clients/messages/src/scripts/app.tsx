import * as React from 'react';
import axios from 'axios';
import UserProfile from './components/user-profile';
import Search from './components/search';
import Chat from './components/chat';
import FloatingActionButton from './components/floating-action-button';
import NewChannelModal from './components/new-channel-modal';
import EditChannelModal from './components/edit-channel-modal';
import DeleteChannelModal from './components/delete-channel-modal';

class App extends React.Component<any, any> {
    constructor(props, context) {
        super(props, context);
        this.state = {
            user: {},
            hasUser: false,
            option: 'chat',
            sessionToken: '',
            channels: [],
            currentChannelIndex: 0,
            messages: new Map(),
            deletedMessageChannel: 0,
            modal: '',
            overlay: false
        };
    }

    public render() {
        if (!this.state.hasUser) {
            return null;
        }
        return (
            <div className="container">
                <div
                    className={this.state.overlay ? 'overlay active' : 'overlay'}
                    onClick={() => this.closeModal()}
                />
                <aside>
                    <h3 onClick={e => this.handleClickMenuOption('default')}>{`Hello, ${
                        this.state.user.firstName
                    }!`}</h3>
                    <nav>
                        <ul>
                            <li
                                className={this.state.option === 'profile' ? 'active' : ''}
                                onClick={e => this.handleClickMenuOption('profile')}
                            >
                                Profile & Account
                            </li>
                            <li
                                className={this.state.option === 'search' ? 'active' : ''}
                                onClick={e => this.handleClickMenuOption('search')}
                            >
                                Search Users
                            </li>
                            <li
                                className={this.state.option === 'chat' ? 'active' : ''}
                                onClick={e => this.handleClickMenuOption('chat')}
                            >
                                Chat
                            </li>
                        </ul>
                    </nav>
                    <div className="log-out" onClick={e => this.signOut(e)}>
                        Sign Out
                        <i className="fa fa-sign-out" aria-hidden="true" />
                    </div>
                </aside>
                <section className="main-panel">{this.renderMainPanel()}</section>
            </div>
        );
    }

    componentWillMount() {
        this.authenticateUser();
        this.fetchChannels();
    }

    private getCurrentHost = (): string => {
        let host: string;
        if (window.location.hostname === 'info-344.zicodeng.me') {
            host = 'info-344-api.zicodeng.me';
        } else {
            host = 'localhost';
        }
        return host;
    };

    private authenticateUser(): void {
        const sessionToken = this.getSessionToken();
        this.setState({
            sessionToken: sessionToken
        });
        const host = this.getCurrentHost();
        const url = `https://${host}/v1/users/me`;

        // Validate this session token.
        fetch(url, {
            method: 'get',
            mode: 'cors',
            headers: new Headers({
                Authorization: sessionToken
            })
        })
            .then(res => {
                if (res.status < 300) {
                    this.establishWebsocket(host);
                    return res.json();
                }
                return res.text();
            })
            .then(data => {
                if (typeof data === 'string') {
                    throw Error(data);
                } else {
                    this.setState({
                        user: data,
                        hasUser: true
                    });
                }
            })
            .catch(error => {
                window.alert(error.message);
                window.location.replace('index.html');
            });
    }

    private establishWebsocket = (host: string): void => {
        const websocket = new WebSocket(`wss://${host}/v1/ws?auth=${this.state.sessionToken}`);
        websocket.addEventListener('error', function(err) {
            console.log(err);
        });
        websocket.addEventListener('open', function() {
            console.log('Websocket connection established');
        });
        websocket.addEventListener('close', function() {
            console.log('Websocket connection closed');
        });
        websocket.addEventListener(
            'message',
            function(event) {
                const messages = this.state.messages;
                const channels = this.state.channels;
                const user = this.state.user;

                let deletedMessageChannel = this.state.deletedMessageChannel;
                let currentChannelIndex = this.state.currentChannelIndex;

                const data = JSON.parse(event.data);
                switch (data.type) {
                    case 'message-new':
                        if (messages.has(data.message.channelID)) {
                            messages.get(data.message.channelID).push(data.message);
                        }
                        break;

                    case 'message-update':
                        if (messages.has(data.message.channelID)) {
                            messages.get(data.message.channelID).map((message, i) => {
                                if (message._id === data.message._id) {
                                    messages.get(data.message.channelID)[i] = data.message;
                                }
                            });
                        }
                        break;

                    case 'message-delete':
                        messages.get(channels[currentChannelIndex]._id).map((message, i) => {
                            if (message._id === data.messageID) {
                                messages.get(channels[currentChannelIndex]._id).splice(i, 1);
                            }
                        });
                        break;

                    case 'channel-new':
                        channels.push(data.channel);
                        messages.set(data.channel._id, []);
                        // If the current user created a new channel,
                        // redirect to the new channel chat page.
                        if (this.state.user.id === data.channel.creator.id) {
                            currentChannelIndex = channels.length - 1;
                        }
                        break;

                    case 'channel-update':
                        channels.map((channel, i) => {
                            if (channel._id === data.channel._id) {
                                channels[i] = data.channel;
                            }
                        });
                        break;

                    case 'channel-delete':
                        const deletedChannelID = data.channelID;

                        // Remove the deleted channel from our local channels state.
                        channels.map((channel, i) => {
                            if (channel._id === deletedChannelID) {
                                const creatorID = channel.creator.id;

                                // If other users are currently on this deleted channel,
                                // prompt a message to inform them
                                // and force them to fallback to default channel.
                                if (deletedChannelID === channels[currentChannelIndex]._id) {
                                    currentChannelIndex = 0;
                                    if (creatorID !== user.id) {
                                        window.alert(
                                            'This channel was just deleted by the channel creator.'
                                        );
                                    }
                                }

                                channels.splice(i, 1);
                                // Delete all messages in this channel.
                                messages.delete(deletedChannelID);
                            }
                        });

                        break;

                    default:
                        break;
                }
                // Update view.
                this.setState({
                    messages: messages,
                    channels: channels,
                    currentChannelIndex: currentChannelIndex
                });
            }.bind(this)
        );
    };

    // Sign out the user and end the session.
    private signOut(e): void {
        e.preventDefault();

        const sessionToken = this.getSessionToken();
        const url = `https://${this.getCurrentHost()}/v1/sessions/mine`;

        fetch(url, {
            method: 'delete',
            mode: 'cors',
            headers: new Headers({
                Authorization: sessionToken
            })
        })
            .then(res => {
                // If the response is successful,
                // remove session token in local storage.
                if (res.status < 300) {
                    this.setState({
                        hasUser: false
                    });
                    localStorage.removeItem('session-token');
                    window.location.replace('index.html');
                }
                return res.text();
            })
            .then(data => {
                if (typeof data === 'string') {
                    throw Error(data);
                }
            })
            .catch(error => {
                window.alert(error.message);
            });
    }

    // Get session token from local storage.
    private getSessionToken(): String | null {
        const sessionToken = localStorage.getItem('session-token');
        if (sessionToken == null || sessionToken.length === 0) {
            // If no session token found in local storage,
            // redirect the user back to landing page.
            window.location.replace('index.html');
        }
        return sessionToken;
    }

    private handleClickMenuOption(option: String): void {
        this.setState({
            option: option
        });
    }

    private fetchChannels = (): void => {
        const url = `https://${this.getCurrentHost()}/v1/channels`;
        axios
            .get(url, {
                headers: {
                    Authorization: this.getSessionToken()
                }
            })
            .then(res => {
                const fetchedChannels = res.data;
                // Only populate default channel first.
                // Other channels will be populated as the user selects
                // different channels.
                this.fetchMessages(fetchedChannels[0]._id);
                this.setState({
                    channels: fetchedChannels,
                    // Set channel general as default.
                    currentChannelIndex: 0
                });
            })
            .catch(error => {
                window.alert(error.response.data);
            });
    };

    private fetchMessages = (channelID): void => {
        const url = `https://${this.getCurrentHost()}/v1/channels/${channelID}`;
        axios
            .get(url, {
                headers: {
                    Authorization: this.getSessionToken()
                }
            })
            .then(res => {
                const messages = this.state.messages;
                messages.set(channelID, res.data);
                this.setState({
                    messages: messages
                });
            })
            .catch(error => {
                window.alert(error.response.data);
            });
    };

    // Render main panel view based on selected menu options.
    renderMainPanel = (): JSX.Element => {
        switch (this.state.option) {
            case 'profile':
                return (
                    <UserProfile
                        user={this.state.user}
                        getSessionToken={this.getSessionToken}
                        updateUser={updatedUser => this.updateUser(updatedUser)}
                    />
                );

            case 'search':
                return <Search sessionToken={this.state.sessionToken} />;

            case 'chat':
                return (
                    <div className="chat-container">
                        {this.renderModal()}
                        {this.getCurrentChannel() ? (
                            <Chat
                                host={this.getCurrentHost()}
                                sessionToken={this.getSessionToken()}
                                currentChannel={this.getCurrentChannel()}
                                messages={this.state.messages}
                                user={this.state.user}
                                openModal={modal => this.openModal(modal)}
                            />
                        ) : null}
                        <FloatingActionButton
                            openModal={modal => this.openModal(modal)}
                            channels={this.state.channels}
                            getCurrentChannelIndex={currentChannelIndex =>
                                this.getCurrentChannelIndex(currentChannelIndex)
                            }
                        />
                    </div>
                );

            default:
                return <h1>{`Howdy, ${this.state.user.firstName}!`}</h1>;
        }
    };

    private renderModal = () => {
        const modal = this.state.modal;
        switch (modal) {
            case 'NewChannel':
                return (
                    <NewChannelModal
                        host={this.getCurrentHost()}
                        sessionToken={this.getSessionToken()}
                        closeModal={() => this.closeModal()}
                    />
                );

            case 'EditChannel':
                return (
                    <EditChannelModal
                        currentChannel={this.getCurrentChannel()}
                        host={this.getCurrentHost()}
                        sessionToken={this.getSessionToken()}
                        closeModal={() => this.closeModal()}
                    />
                );
            case 'DeleteChannel':
                return (
                    <DeleteChannelModal
                        currentChannel={this.getCurrentChannel()}
                        host={this.getCurrentHost()}
                        sessionToken={this.getSessionToken()}
                        closeModal={() => this.closeModal()}
                    />
                );
            default:
                return null;
        }
    };

    // Update user info.
    private updateUser = (updatedUser: Object) => {
        this.setState({
            user: updatedUser
        });
    };

    private getCurrentChannelIndex = (currentChannelIndex): void => {
        const channels = this.state.channels;
        const messages = this.state.messages;
        this.fetchMessages(channels[currentChannelIndex]._id);
        this.setState({
            currentChannelIndex: currentChannelIndex
        });
    };

    private getCurrentChannel = () => {
        const channels = this.state.channels;
        const currentChannelIndex = this.state.currentChannelIndex;
        const currentChannel = channels[currentChannelIndex];
        return currentChannel;
    };

    private openModal = modal => {
        this.setState({
            modal: modal,
            overlay: true
        });
    };

    private closeModal = () => {
        this.setState({
            modal: false,
            overlay: false
        });
    };
}

export default App;
