
# RBACBot -- Role Based Access Control BOT 

> For monorepo users/lovers, this helps you manage write access (through Pull Request) per file/directory basis.

### Usage

```
RBACBOT_WEBHOOK_SECRET=*** RBACBOT_ACCESS_TOKEN=*** rbacbot github
```

Once PRs are open, this will check the config and diff patch and find whom are in charge of these changes (configured via `.rbacbot.yaml` below).

For each diff, only those who met all access rules of all changed files can trigger a `Review Aceept` check, once `numAcceptRequired` is met, the PR is automatically merged.

### Notice

Please make sure the webhook can receive `Pull requests` and `Pull request reviews` events and `repo` scopes for token.


### Environment Variables

|              Name             |              Default          |       Usage                    
|:------------------------------|:-----------------------------:|:------------------------------
|     RBACBOT_CONFIG_NAME       |       .rbacbot.yaml           | config file name in repo        
|         RBACBOT_ADDR          |             :4567             | on which address webhook listens on
|         RBACBOT_TTL           |           1h                  | config/diff patches in memory cache period
|     RBACBOT_WEBHOOK_SECRET    |                               | web hook secret
|      RBACBOT_ACCESS_TOKEN     |                               | oauth/personal access token for github api
 
 
 
 ### .rbacbot.yaml
 
 ```yaml
settings:
  numAcceptRequired: 2
  mergeMethod: squash # merge, rebase
roles:
  supremeleader:   # role name
    - user: joesonw  # user name
access:
  '*':  # file glob
    - supremeleader  # role name
 ```
 
