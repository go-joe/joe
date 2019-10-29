+++
title = "Deploying via Docker"
slug = "docker"
weight = 8
+++

{{% notice note %}}
Like all recipes, this tutorial only shows one way to do things. 
If you think you can improve the example please open an issue or pull request at
[our GitHub repository](https://github.com/go-joe/joe/issues).
{{% /notice %}}

At some point you will want to deploy your Bot to a server so it can run 24/7.
Chances are you already deploy other services using Docker so it makes sense to
also bundle and deploy Joe in that way. This recipe gives you some advice on how
that can be done but there is typically more than one way to achieve this.

Let's start with a simple `Dockerfile`:

```docker
# We use busybox as base image to create very small Docker images.
# Before you use this, you should check if there is a newer version you want to use.
FROM busybox:1.31.1

# We add the minmal dependencies the Go runtime requires to be able to run your
# binary. Depending on your code you might not even have to use those but we
# include them for completeness.
ADD build/zoneinfo.zip /usr/local/go/lib/time/zoneinfo.zip
ADD build/*.crt /etc/ssl/certs/

# Finally we add the binary that contains your bot. Here we called it "my-bot"
# but you can choose the name freely as long as you update it in all three places
# below as well as in the Makefile below.
ADD build/my-bot /bin/my-bot

ENTRYPOINT ["my-bot"]
```

To build this image we create the following `Makefile`:

```makefile
SHELL=/bin/bash

IMAGE=example.com/my-bot # replace with your own image name
VERSION=0.42 # Your image version. Make sure to update this to deploy a new version

.PHONY: build
build:
	mkdir -p build
	cp "$$GOROOT/lib/time/zoneinfo.zip" build/zoneinfo.zip
	cp /etc/ssl/certs/*.crt ./build/
	CGO_ENABLED=0 go build -v -tags netgo -ldflags "-extldflags '-static' -w" -o build/my-bot
	docker build -t $(IMAGE):$(VERSION) .
	rm -Rf build

.PHONY: push
push:
	docker push $(IMAGE):$(VERSION)
```

This Makefile should be put next to your code. It will create a temporary `build`
directory, copy over timezone and certificate files and then statically build your
bot.

The image can be built and pushed in a single command via `make build push`.

Finally if you want to deploy the Bot to Kubernetes you can use the following Manifest:

```yaml
kind: Deployment
apiVersion: apps/v1beta1
metadata:
  name: my-bot
spec:
  replicas: 1
  strategy:
    type: Recreate
  template:
    metadata: { labels: { app: my-bot } }
    spec:
      containers:
        - name:  bot
          image: example.com/my-bot:0.42 # Make sure to update the image name and version according to your Makefile
          imagePullPolicy: Always
          ports:
            - containerPort: 80
          env:
            - name:  "SLACK_TOKEN"
              value: "…"
            - name:  "HTTP"
              value: ":80"
            # Add other environment settings you may need. Consider using Kubernetes secrets for the credentials. 
---

kind: Service
apiVersion: v1
metadata: { name: my-bot }
spec:
  selector: { app: my-bot }
  ports:
    - port: 80

---

kind: Ingress
apiVersion: extensions/v1beta1
metadata:
  name: my-bot
spec:
  rules:
    - host: my-bot.example.com
      http:
        paths:
          - backend: { serviceName: my-bot, servicePort: 80}
```
