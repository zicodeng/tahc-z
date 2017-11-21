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
            option: 'default',
            sessionToken: '',
            channels: [],
            selectedChannel: {},
            messages: [],
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
                    <h3 onClick={e => this.handleClickMenuOption('default')}>{`Hello, ${this.state
                        .user.firstName}!`}</h3>
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
                let messages = this.state.messages;
                let channels = this.state.channels;
                let selectedChannel = this.state.selectedChannel;
                const data = JSON.parse(event.data);
                switch (data.type) {
                    case 'message-new':
                        messages.push(data.message);
                        break;

                    case 'message-update':
                        messages.map((message, i) => {
                            if (message._id == data.message._id) {
                                messages[i] = data.message;
                            }
                        });
                        break;

                    case 'message-delete':
                        messages.map((message, i) => {
                            if (message._id == data.message) {
                                messages.splice(i, 1);
                            }
                        });
                        break;

                    case 'channel-new':
                        channels.push(data.channel);
                        // Set the currently selected channel to the newly created channel.
                        selectedChannel = data.channel;
                        break;

                    case 'channel-update':
                        selectedChannel = data.channel;
                        channels.map((channel, i) => {
                            if (channel._id === selectedChannel._id) {
                                channels[i] = selectedChannel;
                            }
                        });
                        break;

                    case 'channel-delete':
                        const deletedChannelID = data.channel;
                        // Fallback to default channel;
                        selectedChannel = channels[0];
                        // Remove the deleted channel from our local channels state.
                        channels.map((channel, i) => {
                            if (channel._id === deletedChannelID) {
                                channels.splice(i, 1);
                            }
                        });
                        // Delete all messages in our local state.
                        messages = [];
                        break;
                    default:
                        break;
                }
                // Update view.
                this.setState({
                    messages: messages,
                    channels: channels,
                    selectedChannel: selectedChannel
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
                this.setState({
                    channels: fetchedChannels,
                    // Set channel general as default.
                    selectedChannel: fetchedChannels[0]
                });
                this.fetchMessages(fetchedChannels[0]._id);
            })
            .catch(error => {
                console.log(error.response.data);
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
                this.setState({
                    messages: res.data
                });
            })
            .catch(error => {
                console.log(error.response.data);
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
                        <Chat
                            host={this.getCurrentHost()}
                            sessionToken={this.getSessionToken()}
                            channel={this.state.selectedChannel}
                            messages={this.state.messages}
                            user={this.state.user}
                            openModal={modal => this.openModal(modal)}
                        />
                        <FloatingActionButton
                            openModal={modal => this.openModal(modal)}
                            channels={this.state.channels}
                            getSelectedChannel={selectedChannel =>
                                this.getSelectedChannel(selectedChannel)}
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
                        selectedChannel={this.state.selectedChannel}
                        host={this.getCurrentHost()}
                        sessionToken={this.getSessionToken()}
                        closeModal={() => this.closeModal()}
                    />
                );
            case 'DeleteChannel':
                return (
                    <DeleteChannelModal
                        selectedChannel={this.state.selectedChannel}
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

    private getSelectedChannel = (selectedChannel): void => {
        this.setState({
            selectedChannel: selectedChannel
        });
        this.fetchMessages(selectedChannel._id);
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
