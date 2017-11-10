import * as React from 'react';

class PasswordResetForm extends React.Component<any, any> {
    constructor(props, context) {
        super(props, context);

        this.state = {
            email: '',
            step: 1,
            stepOneStatus: 'InProgress',
            error: ''
        };
    }

    public render() {
        return (
            <div className="password-reset-container">
                <div className="progress-bar">
                    <ul>
                        <li
                            className={
                                this.state.step === 1 || this.state.stepOneStatus === 'Completed'
                                    ? 'active'
                                    : ''
                            }
                            onClick={e => this.handleClickStep(1)}
                        >
                            1
                        </li>
                        <li
                            className={this.state.step === 2 ? 'with-bar active' : 'with-bar'}
                            onClick={e => this.handleClickStep(2)}
                        >
                            2
                        </li>
                    </ul>
                </div>
                <div
                    className="material-form"
                    style={this.state.step === 1 ? {} : { display: 'none' }}
                >
                    <h1 className="title">Request Password Reset Code</h1>
                    <form onSubmit={e => this.handleSubmitRequestResetCode(e)}>
                        <div className="input-container">
                            <input type="email" ref="email" required />
                            <label htmlFor="email">Email</label>
                            <div className="bar" />
                        </div>
                        <div className="button-container">
                            <button>
                                <span>Send Me Code</span>
                            </button>
                        </div>
                    </form>
                    <p className="error">{this.state.step === 1 ? this.state.error : ''}</p>
                </div>
                <div
                    className="material-form"
                    style={this.state.step === 2 ? {} : { display: 'none' }}
                >
                    <h1 className="title">New Password</h1>
                    <form onSubmit={e => this.handleSubmitNewPassword(e)}>
                        <div className="input-container">
                            <input type="password" ref="password" required />
                            <label htmlFor="password">Password</label>
                            <div className="bar" />
                        </div>
                        <div className="input-container">
                            <input type="password" ref="passwordConf" required />
                            <label htmlFor="password-conf">Confirm Your Password</label>
                            <div className="bar" />
                        </div>
                        <div className="input-container">
                            <input type="text" ref="resetCode" required />
                            <label htmlFor="reset-code">Reset Code</label>
                            <div className="bar" />
                        </div>
                        <div className="button-container">
                            <button>
                                <span>Save</span>
                            </button>
                        </div>
                    </form>
                    <p className="error">{this.state.step === 2 ? this.state.error : ''}</p>
                </div>
                <p className="back" onClick={e => this.handleClickBack()}>
                    Back to Sign-In/Sign-Up
                </p>
            </div>
        );
    }

    private handleSubmitRequestResetCode = (e): void => {
        e.preventDefault();
        const email = this.refs.email['value'];
        const resetCodeRequest = {
            email: email
        };
        this.setState({
            email: email
        });

        let url;
        if (window.location.hostname === 'info-344.zicodeng.me') {
            url = 'https://info-344-api.zicodeng.me/v1/resetcodes';
        } else {
            url = 'https://localhost/v1/resetcodes';
        }

        let ok = false;
        fetch(url, {
            method: 'post',
            body: JSON.stringify(resetCodeRequest),
            mode: 'cors',
            headers: new Headers({
                'Content-Type': 'application/json'
            })
        })
            .then(res => {
                ok = res.ok;
                return res.text();
            })
            .then(text => {
                // If no HTTP error is responded from the server,
                // go to step 2.
                if (ok) {
                    this.setState({
                        step: 2,
                        stepOneStatus: 'Completed',
                        error: ''
                    });
                    window.alert(text);
                } else {
                    this.setState({
                        error: text
                    });
                }
            })
            .catch(error => {
                console.log(error);
            });
    };

    private handleSubmitNewPassword = (e): void => {
        e.preventDefault();

        const password = this.refs.password['value'];
        const passwordConf = this.refs.passwordConf['value'];
        const resetCode = this.refs.resetCode['value'];
        const newPassword = {
            password: password,
            passwordConf: passwordConf,
            resetCode: resetCode
        };

        let url;
        if (window.location.hostname === 'info-344.zicodeng.me') {
            url = 'https://info-344-api.zicodeng.me/v1/passwords?email=' + this.state.email;
        } else {
            url = 'https://localhost/v1/passwords?email=' + this.state.email;
        }

        fetch(url, {
            method: 'put',
            body: JSON.stringify(newPassword),
            mode: 'cors',
            headers: new Headers({
                'Content-Type': 'application/json'
            })
        })
            .then(res => {
                if (res.ok) {
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
                    this.refs.password['value'] = '';
                    this.refs.passwordConf['value'] = '';
                    this.refs.resetCode['value'] = '';
                    this.setState({
                        error: ''
                    });
                    window.location.replace('app.html');
                }
            })
            .catch(error => {
                this.setState({
                    error: error.message
                });
            });
    };

    private handleClickStep(step: number): void {
        if (step === 1) {
            this.setState({
                step: step
            });
        } else {
            this.setState({
                step: step,
                stepOneStatus: 'Completed'
            });
        }
    }

    // Return back to sign-in/sign-up view.
    private handleClickBack = (): void => {
        this.props.switchForm('Signing');
    };
}

export default PasswordResetForm;
