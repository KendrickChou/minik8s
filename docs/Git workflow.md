# Git Workflow

## Syncing

* ``git remote``: 本质是一个书签，提供url的别名

  ```
  git remote add <name> <url>
  git remote rm <name>
  git remote rename <old-name> <new-name>
  ```

  在``git clone``时，会自动创建origin到cloned repo的connection。

* ``git fetch``: 用来拉取远端的内容。

  不会把拉下来的内容和本地内容强制merge，而是把他们隔离开，可以用 ``git checkout``来显式查看。

  ``git fetch`` VS ``git pull``: 

  1. ``git fetch`` is the 'safe' version. ·`git pull` is more aggressive.
  2. ``git fetch``下载远端内容，但不会更新local repo
  3. `git pull`下载远端内容并会立刻执行`git merge`

  ```
  git fetch <remote>
  git fetch <remote> <branch>
  git fetch --all
  git fetch --dry-run  // It will output examples of actions it will take during the fetch but not apply them.
  ```

* `git push`

  Amended force push: 用来在commit的时候修改commit message 或者 add new changes.

  ```
  # make changes to a repo and git add
  git commit --amend
  # update the existing commit message
  git push --force origin main
  ```

* `git pull`: 多使用rebase来避免不必要的merge！

  example:

  <img src=".\pics\git pull example - before.png" alt="git pull example - before" style="zoom:67%;" />

  直接`git pull`:

  <img src=".\pics\git pull example - after.png" alt="git pull example - after" style="zoom:60%;" />

  `git pull --rebase`: 重新修改E,F,G 让他们变成线性。

  <img src=".\pics\git pull example - rebase after.png" alt="git pull example - rebase after" style="zoom:60%;" />

  Pulling via Rebase: 可以让记录变成线性的，**避免不必要的merge**。

## Making a Pull Request

* Base function: 用来提醒所有开发者完成了一个功能

* When you file a pull request, all you’re doing is *requesting* that another developer (e.g., the project maintainer) *pulls* a branch from your repository into their repository.  This means that you need to provide 4 pieces of information to file a pull request: 

  1. the source repository
  2.  the source branch
  3.  the destination repository
  4.  the destination branch.

* workflow:

  1. A developer creates the feature in a dedicated branch in their local repo.
  2. The developer pushes the branch to a public repository.
  3. The developer files a pull request.
  4. The rest of the team reviews the code, discusses it, and alters it.
  5. The project maintainer merges the feature into the official repository and closes the pull request.

* 具体例子详见：

  https://www.atlassian.com/git/tutorials/making-a-pull-request

## Branches

* branches are just pointers to commits. When you create a branch, all Git needs to do is create a new pointer, it doesn’t change the repository in any other way.

* ``git merge``: `git merge` takes two commit pointers, usually the branch tips, and will find a common base commit between them. Once Git finds a common base commit it will create a new "merge commit" that combines the changes of each queued merge commit sequence.

  **If your feature branch was actually as small as the one in the above example, you would probably be better off rebasing it onto `main` and doing a fast-forward merge. This prevents superfluous merge commits from cluttering up the project history.**

  **Explicit Merge** vs **Implicit Merge**：

  1. explicit 会产生merge commit，在大的feature合并的时候用
  2. implicit 是基于rebase的，在修复小bug或做一些小工作的时候用。

## Comparing Workflows

* Centralized Workflow
* Feature Branch Workflow
* Gitflow Workflow :white_check_mark:
* Forking Workflow

