import * as React from 'react';

import SigningForm from './components/signing-form';
import PasswordResetForm from './components/password-reset-form';

class Index extends React.Component<any, any> {
    private count: number = 0;

    constructor(props, context) {
        super(props, context);

        this.state = {
            form: 'Signing'
        };
    }

    public render() {
        return (
            <div className="container">
                <div className="brand">
                    <h1>Tahc-Z</h1>
                    <p>A more powerful version of Z-Chat</p>
                </div>
                {this.renderForm()}
            </div>
        );
    }

    private renderForm = () => {
        switch (this.state.form) {
            case 'Signing':
                return <SigningForm switchForm={form => this.switchForm(form)} />;

            case 'PasswordReset':
                return <PasswordResetForm switchForm={form => this.switchForm(form)} />;

            default:
                break;
        }
    };

    private switchForm = (form: string): void => {
        this.setState({
            form: form
        });
    };
}

export default Index;
