# ReferenceGrant POC

This project is a proof of concept meant to show the viability of the next
generation of ReferenceGrant. As a POC, this project provides absolutely no
stability, and should never be used in a production environment. If this ever
becomes production ready, it will do so exclusively within a kubernetes or
kubernetes-sigs repo.

## High Level Goals

* Show how ReferenceGrant could become part of kubernetes/kubernetes via
  sig-auth.
* Enable ReferenceGrant to be used more generically, defining the specific
  reference paths that should be followed.
* Provide a means of authorizing controllers to only access the resources that
  are directly referenced by resources they are implementing. (For example, a
  Gateway controller should only be reading from the secrets referenced by a
  Gateway).
* Provide the foundation for a backfill that could be used to provide similar
  functionality in earlier Kubernetes versions.

## Context

With SIG-Storage adopting ReferenceGrant for [cross-namespace storage data
sources](https://kubernetes.io/blog/2023/01/02/cross-namespace-data-sources-alpha/),
it became important for us to transition ReferenceGrant to a more neutral home.
This project explores what a transition to a more generic, auth-first approach
could look like.

This has been a point of discussion at previous KubeCons, resulting in both a
[KEP](https://github.com/kubernetes/enhancements/issues/3766) and a [more recent
doc](https://docs.google.com/document/d/1poQb0uxOkJsebNgTMrpaogcY9vcehGHe1myqvenCXtU/edit)
showing how this could all work.
