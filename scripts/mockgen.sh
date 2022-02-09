#!/bin/bash
echo `pwd`
mockgen -destination=../yandex/mocks/mock.go -package=mocks github.com/yandex-cloud/terraform-provider-yandex/yandex $1
