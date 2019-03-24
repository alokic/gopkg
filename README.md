# gopkg
Home for Golang libraries

# Attention - You may break things!
This is a base library package which is included by many golang projects.
Exercise caution while pushing changes here:
1. Make sure drone builds are ok.
2. Codeclimate is clean.
3. If introducing a new external library, think whether you need to do semantic versioning. If not then app importing these pkgs may pick up a version different from the one with which library is tested.
4. Use release in your dep managemnet tool rather than picking up `master` branch.
