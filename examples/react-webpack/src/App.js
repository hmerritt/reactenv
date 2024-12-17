const App = () => (
	<div>
		<h1>reactenv</h1>
		<p>
			REACT_APP_VAR: <strong>{process.env.REACT_APP_VAR}</strong>
		</p>
		<p>
			REACT_APP_API_URL: <strong>{process.env.REACT_APP_API_URL}</strong>
		</p>
	</div>
);

export default App;
