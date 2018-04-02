# Command Type: Start

This command should be sent from the calling program immediately

```
{
  "type": "start",

  "featuresConfig": {

    // array of paths to features that need to be loaded
    "absolute_paths": [],

    // filters to select specific scenarios to run
    "filters": {

      // array of strings, which will become regular expressions, that a scenario name must match
      // if empty, all features match
      // if multiple names provided, a scenario needs to match only one
      "names": [],

      // tag expression for what scenarios should run on
      "tag_expression": "",

      // map from feature path to array of line numbers for what scenarios to run
      "lines": {
        "/path/to/feature": [1],
        //...
      },
    }
  },

  "runtimeConfig": {

    // if true, after the first failure, the remaining scenarios are skipped
    "is_fail_fast": false,

    // if true, do not run any steps
    "is_dry_run": false,

    // if true, pending steps cause the test run to fail
    "is_strict": false,

    // hooks to run before each test case
    "before_test_case_hook_definitions": [
      {
        // a unique id for the before hook
        "id": "",

        // tag expression for what scenarios this hook should run on
        "tag_expression": "",

        // uri / line for where the hook was defined (optional)
        "uri": "",
        "line": ""
      }
      // ...
    ],

    // hooks to run after each test case (same format as before_test_case_hook_definitions)
    "after_test_case_hook_definitions": [],

    "step_definitions": [
      {

        // a unique id for the step
        "id": "",

        "pattern": {
          // text or regexp as string
          "source": "",
          // "regular_expression" or "cucumber_expression"
          "type": ""
        },

        // uri / line for where the hook was defined (optional)
        "uri": "",
        "line": ""
      }
      // ...
    ],

    "parameterTypes": [
      {
        // a unique name for the parameter type
        "name": "",

        // array of regexp sources
        "regexps": [],
      }
      // ...
    ]
  }
}
```