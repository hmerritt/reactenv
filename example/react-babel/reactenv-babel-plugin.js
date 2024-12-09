const { types: t } = require("@babel/core");

const envBabelPlugin = () => {
    return {
        name: "reactenv-transform-env-variables",
        visitor: {
            MemberExpression(path) {
                // Skip transformation if not in production
                if (process.env.NODE_ENV !== "production") return;

                // Check if we're accessing process.env
                if (
                    path.get("object").matchesPattern("process.env") ||
                    (path.get("object").isMemberExpression() &&
                        path.get("object.object").isIdentifier({ name: "process" }) &&
                        path.get("object.property").isIdentifier({ name: "env" }))
                ) {
                    // Get the environment variable name
                    const envVarName = path.get("property").node.name;

                    // Replace process.env.X with "reactenv.X"
                    path.replaceWith(t.stringLiteral(`reactenv.${envVarName}`));
                }
            },
        },
    };
};

module.exports = envBabelPlugin;
