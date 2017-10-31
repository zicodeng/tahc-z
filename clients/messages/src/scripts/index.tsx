import * as React from 'react';

import LoginForm from './components/login-form';

class Index extends React.Component<any, any> {
    private count: number = 0;

    constructor(props, context) {
        super(props, context);
    }

    public render() {
        return (
            <div className="container">
                <div className="brand">
                    <h1>Tahc-Z</h1>
                    <p>A more powerful version of Z-Chat</p>
                </div>
                <LoginForm />
            </div>
        );
    }
}

export default Index;
