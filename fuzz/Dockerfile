# Copyright 2024 VirtEngine Authors
# 
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# OSS-Fuzz Dockerfile for VirtEngine
# This builds the fuzzing infrastructure for continuous fuzzing via OSS-Fuzz
#
# To test locally:
#   docker build -f fuzz/Dockerfile -t virtengine-fuzz .
#   docker run --rm -v $(pwd)/fuzz/corpus:/corpus virtengine-fuzz
#
# Task Reference: QUALITY-002 - Fuzz Testing Implementation

FROM gcr.io/oss-fuzz-base/base-builder-go

# Install additional dependencies
RUN apt-get update && apt-get install -y \
    libusb-1.0-0-dev \
    libudev-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy source code
COPY . $SRC/virtengine
WORKDIR $SRC/virtengine

# Build script will be invoked by OSS-Fuzz infrastructure
COPY fuzz/build.sh $SRC/
