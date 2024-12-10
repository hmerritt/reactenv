const App = () => (
    <div>
        <h1>reactenv</h1>
        <p>
            REACT_APP_SECRET: <strong>{import.meta.env.VITE_SECRET}</strong>
        </p>
        <p>
            REACT_APP_API_URL: <strong>{import.meta.env.VITE_API_URL}</strong>
        </p>
    </div>
);

export default App;
