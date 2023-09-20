[![Board Status](https://dev.azure.com/ALMP-ORG-P01/4beb804a-d723-49ea-b9e1-bc360124b3e7/90f09c14-11be-494b-b116-d725ee917424/_apis/work/boardbadge/20805e31-098d-4933-b907-e706bcb6b0cd)](https://dev.azure.com/ALMP-ORG-P01/4beb804a-d723-49ea-b9e1-bc360124b3e7/_boards/board/t/90f09c14-11be-494b-b116-d725ee917424/Philips.PlanningCategory)
# go-pjlink-emulator

pjlink device emulator written in GO

## Implemented features:

- Power on/off
- Name
- Lamp life time (Simulated)
- Input switching
- Class query of protocol version

## Fixes:

- Keep device state after socket close.

## Working setup

Emulator is validated against, PJLink Class 1 and 2:
https://pjlink.jbmia.or.jp/english/index.html

Software PJLink Test tool version 2.0.1.0.190508:

PJLink pdf spec. class 2:
https://pjlink.jbmia.or.jp/english/data_cl2/PJLink_5-1.pdf

Download package:
https://pjlink.jbmia.or.jp/english/data_cl2/PJLink_5-2.zip
