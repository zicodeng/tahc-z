import * as React from 'react';

class App extends React.Component<any, any> {
    private count: number = 0;

    constructor(props, context) {
        super(props, context);
    }

    public render() {
        return (
            <div>
                <h1>Tahc-Z</h1>
            </div>
        );
    }
}

export default App;
