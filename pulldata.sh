#!/bin/bash

ITEMS="Lnl_BadgeType Lnl_BadgeStatus Lnl_Badge Lnl_Cardholder Lnl_AccessLevel Lnl_AccessLevelAssignment"

OLDIFS=$IFS
IFS=" "

DIR=$1
CONFIG=$2

if [ "$DIR" = "" ]; then
  echo "No target directory specified!"
  exit 1
fi

if [ "$CONFIG" = "" ]; then
  echo "No config specified!"
  exit 1
fi

mkdir -p "$DIR"

for ITEM in $ITEMS
do
	cmd/querytool/querytool -c "$CONFIG" -t $ITEM -f "${DIR}/${ITEM}.json"
done

IFS=$OLDIFS
