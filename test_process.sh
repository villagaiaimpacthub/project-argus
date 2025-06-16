#!/bin/bash
echo "Test process started"
echo "Normal output line 1"
echo "Error: This is a test error" >&2
echo "Normal output line 2"
echo "Failed to do something" >&2
echo "Process completed"
sleep 2
echo "Final output"
