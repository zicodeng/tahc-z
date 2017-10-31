import * as React from 'react';
import UserProfile from './components/user-profile';

class App extends React.Component<any, any> {
    constructor(props, context) {
        super(props, context);
        this.state = {
            user: {},
            hasUser: false,
            option: ''
        };
    }

    public render() {
        if (!this.state.hasUser) {
            return <div />;
        }
        return (
            <div className="container">
                <aside>
                    <h3>{`Hello, ${this.state.user.firstName}!`}</h3>
                    <nav>
                        <ul>
                            <li
                                className={this.state.option === 'profile' ? 'active' : ''}
                                onClick={e => this.handleClickMenuOption('profile')}
                            >
                                Profile & Account
                            </li>
                        </ul>
                    </nav>
                    <div className="log-out" onClick={e => this.logOut(e)}>
                        Log Out
                        <i className="fa fa-sign-out" aria-hidden="true" />
                    </div>
                </aside>
                <section className="main-panel">{this.renderMainPanel()}</section>
            </div>
        );
    }

    componentWillMount() {
        this.authenticateUser();
    }

    private authenticateUser(): void {
        const sessionToken = this.getSessionToken();

        let url;
        if (window.location.hostname === 'info-344.zicodeng.me') {
            url = 'https://info-344-api.zicodeng.me/v1/users/me';
        } else {
            url = 'https://localhost/v1/users/me';
        }

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

    // Log out the user and end the session.
    private logOut(e): void {
        e.preventDefault();

        const sessionToken = this.getSessionToken();

        let url;
        if (window.location.hostname === 'info-344.zicodeng.me') {
            url = 'https://info-344-api.zicodeng.me/v1/sessions/mine';
        } else {
            url = 'https://localhost/v1/sessions/mine';
        }
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

    // Render main panel view based on selected menu options.
    private renderMainPanel() {
        switch (this.state.option) {
            case 'profile':
                return (
                    <UserProfile
                        user={this.state.user}
                        getSessionToken={this.getSessionToken}
                        updateUser={updatedUser => this.updateUser(updatedUser)}
                    />
                );

            default:
                return <h1>{`Howdy, ${this.state.user.firstName}!`}</h1>;
        }
    }

    // Update user info.
    private updateUser = (updatedUser: Object) => {
        this.setState({
            user: updatedUser
        });
    };
}

export default App;
