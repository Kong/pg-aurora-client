## Purpose of checklist is to act as a reminder. Delete checklist once you are satisfied.

### Development Checklist
- [ ] For Koko OSS , followed [repo guidelines](https://konghq.atlassian.net/wiki/spaces/TK/pages/2737930261/Repository+management#koko).
- [ ] Did you follow [commit message rules](https://konghq.atlassian.net/wiki/spaces/TK/pages/2737930261/Repository+management#Commit-message-format)
- [ ] Did you follow [branch naming rules](https://konghq.atlassian.net/wiki/spaces/TK/pages/2737930261/Repository+management#%5BhardBreak%5DNaming-conventions)

### Deployment Checklist
- [ ] Dev
- [ ] Prod

**If you checked dev**
- [ ] Did you add feature flags within your code/helm to make sure that this does not get rolled out to prod by accident?
- [ ] Did you make changes to values.yaml and add flag-off tags to values-cloud-02-prod.yaml?

**If you checked prod**
- [ ] Did you deploy this and dev and does it work?
- [ ] Did you create a merge commit from deployment-dev to deployment-prod and modify values-cloud-02-prod.yaml with the right release image tag?

**Common Preflight checks**

External Dependencies

- [ ] Do they exist?
- [ ] Are they reachable from your deployment environment?
- [ ] Do they require credentials? Have you used the credentials and do they work?
- [ ] Are they are trust-worthy?

Deployment
- [ ] If you make Helm/values changes, did you run helm template?
- [ ] Do you have log statements at appropriate level to confirm whether your code is actually running?
- [ ] Do you have tests (synthetic or manual or logging) to validate success or failure?
- [ ] Did you make breaking changes to the data-model/storage so that your code cannot be rolled-back?
- [ ] What is your roll-back plan?


### Congratulations! You made it here! Put your actual commit message and delete the above.