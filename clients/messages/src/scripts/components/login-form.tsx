import * as React from 'react';

class LoginForm extends React.Component<any, any> {
    constructor(props, context) {
        super(props, context);

        this.state = {
            isRegister: false,
            loginError: '',
            registerError: ''
        };
    }

    public render() {
        return (
            <div className={this.state.isRegister ? 'form-container active' : 'form-container'}>
                <div className="card" />
                <div className="card">
                    <h1 className="title">Login</h1>
                    <form onSubmit={e => this.handleSubmitLoginForm(e)}>
                        <div className="input-container">
                            <input type="email" ref="loginEmail" required />
                            <label htmlFor="email">Email</label>
                            <div className="bar" />
                        </div>
                        <div className="input-container">
                            <input type="password" ref="loginPassword" required />
                            <label htmlFor="password">Password</label>
                            <div className="bar" />
                        </div>
                        <div className="button-container">
                            <button>
                                <span>Go</span>
                            </button>
                        </div>
                        <div className="footer">
                            <a href="#">Forgot your password?</a>
                        </div>
                    </form>
                    <p className="error">{this.state.loginError}</p>
                </div>
                <div className="card alt">
                    <div
                        className={this.state.isRegister ? 'toggle active' : 'toggle'}
                        onClick={e => this.openRegisterForm()}
                    />
                    <h1 className="title">
                        Register
                        <div className="close" onClick={e => this.closeRegisterForm()} />
                    </h1>
                    <form onSubmit={e => this.handleSubmitRegisterForm(e)}>
                        <div className="input-container">
                            <input type="text" ref="userName" required />
                            <label htmlFor="username">Username</label>
                            <div className="bar" />
                        </div>
                        <div className="input-container">
                            <input type="text" ref="firstName" required />
                            <label htmlFor="firstname">First Name</label>
                            <div className="bar" />
                        </div>
                        <div className="input-container">
                            <input type="text" ref="lastName" required />
                            <label htmlFor="lastname">Last Name</label>
                            <div className="bar" />
                        </div>
                        <div className="input-container">
                            <input type="email" ref="email" required />
                            <label htmlFor="email">Email</label>
                            <div className="bar" />
                        </div>
                        <div className="input-container">
                            <input type="password" ref="password" required />
                            <label htmlFor="password">Password</label>
                            <div className="bar" />
                        </div>
                        <div className="input-container">
                            <input type="password" ref="passwordConf" required />
                            <label htmlFor="repeat password">Repeat Password</label>
                            <div className="bar" />
                        </div>
                        <div className="button-container">
                            <button>
                                <span>SIGN UP</span>
                            </button>
                        </div>
                    </form>
                    <p className="error">{this.state.registerError}</p>
                </div>
            </div>
        );
    }

    private openRegisterForm(): void {
        this.setState({
            isRegister: true,
            loginError: '',
            registerError: ''
        });
    }

    private closeRegisterForm(): void {
        this.setState({
            isRegister: false,
            loginError: '',
            registerError: ''
        });
    }

    // Sign in the user and save the session token in local storage.
    private handleSubmitLoginForm(e): void {
        e.preventDefault();
        const email = this.refs.loginEmail['value'];
        const password: String = this.refs.loginPassword['value'];

        const userCredential: Object = {
            email: email,
            password: password
        };

        let url;

        if (window.location.hostname === 'info-344.zicodeng.me') {
            url = 'https://info-344-api.zicodeng.me/v1/sessions';
        } else {
            url = 'https://localhost/v1/sessions';
        }

        fetch(url, {
            method: 'post',
            body: JSON.stringify(userCredential),
            mode: 'cors',
            headers: new Headers({
                'Content-Type': 'application/json'
            })
        })
            .then(res => {
                // If we get a successful response (status code < 300),
                // save the contents of the Authorization response header to local storage.
                if (res.status < 300) {
                    // Save session token to local storage.
                    const sessionToken = res.headers.get('Authorization');

                    if (sessionToken != null) {
                        localStorage.setItem('session-token', sessionToken);
                    }
                    return res.json();
                }

                // If response is not ok,
                // catch the error contained in body.
                return res.text();
            })
            .then(data => {
                // If data type is string,
                // it means this is an error sent by server.
                if (typeof data === 'string') {
                    throw Error(data);
                } else {
                    // If the data type is not a string,
                    // it means the user is authenticated,
                    // clear form and redirect the user to app page.
                    this.refs.loginEmail['value'] = '';
                    this.refs.loginPassword['value'] = '';
                    this.setState({
                        loginError: ''
                    });
                    window.location.replace('app.html');
                }
            })
            .catch(error => {
                this.setState({
                    loginError: error.message
                });
            });
    }

    private handleSubmitRegisterForm(e): void {
        e.preventDefault();

        const userName: String = this.refs.userName['value'];
        const firstName: String = this.refs.firstName['value'];
        const lastName: String = this.refs.lastName['value'];
        const email: String = this.refs.email['value'];
        const password: String = this.refs.password['value'];
        const passwordConf: String = this.refs.passwordConf['value'];

        const user: Object = {
            userName: userName,
            lastName: lastName,
            firstName: firstName,
            email: email,
            password: password,
            passwordConf: passwordConf
        };

        let url;

        if (window.location.hostname === 'info-344.zicodeng.me') {
            url = 'https://info-344-api.zicodeng.me/v1/users';
        } else {
            url = 'https://localhost/v1/users';
        }

        fetch(url, {
            method: 'post',
            body: JSON.stringify(user),
            mode: 'cors',
            headers: new Headers({
                'Content-Type': 'application/json'
            })
        })
            .then(res => {
                if (res.status < 300) {
                    // Save session token to local storage.
                    const sessionToken = res.headers.get('Authorization');

                    if (sessionToken != null) {
                        localStorage.setItem('session-token', sessionToken);
                    }
                    return res.json();
                }

                return res.text();
            })
            .then(data => {
                if (typeof data === 'string') {
                    throw Error(data);
                } else {
                    this.refs.userName['value'] = '';
                    this.refs.firstName['value'] = '';
                    this.refs.lastName['value'] = '';
                    this.refs.email['value'] = '';
                    this.refs.password['value'] = '';
                    this.refs.passwordConf['value'] = '';
                    this.setState({
                        registerError: ''
                    });
                    window.location.replace('app.html');
                }
            })
            .catch(error => {
                this.setState({
                    registerError: error.message
                });
            });
    }
}

export default LoginForm;
