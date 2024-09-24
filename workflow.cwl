#!/usr/bin/env cwl-runner

cwlVersion: v1.2
class: Workflow

inputs:
  DT5301: Directory
  DT5302: Directory
  DT5303: Directory

outputs:
  DT5301:
    type: Directory
    outputSource: SS5302/DT5301
  DT5302:
    type: Directory
    outputSource: SS5302/DT5302
  DT5303:
    type: Directory
    outputSource: SS5302/DT5303
  DT5305:
    type: Directory
    outputSource: SS5303/DT5305

steps:
  SS5301:
    in:
      DT5301: DT5301
      DT5302: DT5302
      DT5303: DT5303
    run:
      class: Operation
      inputs:
        DT5301: Directory
        DT5302: Directory
        DT5303: Directory
      outputs:
        DT5304: Directory
    out:
    - DT5304
  SS5302:
    in:
      DT5301: DT5301
      DT5302: DT5302
      DT5303: DT5303
      DT5304: SS5301/DT5304
    run:
      class: Operation
      inputs:
        DT5301: Directory
        DT5302: Directory
        DT5303: Directory
        DT5304: Directory
      outputs:
        DT5301: Directory
        DT5302: Directory
        DT5303: Directory
    out:
    - DT5301
    - DT5302
    - DT5303
  SS5303:
    in:
      DT5301: SS5302/DT5301
      DT5302: SS5302/DT5302
      DT5303: SS5302/DT5303
      DT5304: SS5301/DT5304
    run:
      class: Operation
      inputs:
        DT5301: Directory
        DT5302: Directory
        DT5303: Directory
        DT5304: Directory
      outputs:
        DT5305: Directory
    out:
    - DT5305
