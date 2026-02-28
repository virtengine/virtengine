/**
 * Copyright (c) VirtEngine, Inc.
 * SPDX-License-Identifier: BSL-1.1
 */

package notifications

import "time"

func cloneNotification(notif Notification) Notification {
	cloned := notif
	cloned.Data = cloneStringMap(notif.Data)
	cloned.Channels = append([]Channel(nil), notif.Channels...)
	cloned.ReadAt = cloneTimePtr(notif.ReadAt)
	return cloned
}

func clonePreferences(prefs Preferences) Preferences {
	cloned := prefs
	cloned.Channels = cloneChannelMap(prefs.Channels)
	cloned.Frequencies = cloneFrequencyMap(prefs.Frequencies)
	cloned.QuietHours = cloneQuietHours(prefs.QuietHours)
	return cloned
}

func cloneDeviceToken(device DeviceToken) DeviceToken {
	cloned := device
	cloned.DisabledAt = cloneTimePtr(device.DisabledAt)
	cloned.LastDeliveredAt = cloneTimePtr(device.LastDeliveredAt)
	cloned.LastFailureAt = cloneTimePtr(device.LastFailureAt)
	return cloned
}

func cloneStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	out := make(map[string]string, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneChannelMap(input map[NotificationType][]Channel) map[NotificationType][]Channel {
	if len(input) == 0 {
		return nil
	}
	out := make(map[NotificationType][]Channel, len(input))
	for key, value := range input {
		out[key] = append([]Channel(nil), value...)
	}
	return out
}

func cloneFrequencyMap(input map[NotificationType]DeliveryFrequency) map[NotificationType]DeliveryFrequency {
	if len(input) == 0 {
		return nil
	}
	out := make(map[NotificationType]DeliveryFrequency, len(input))
	for key, value := range input {
		out[key] = value
	}
	return out
}

func cloneQuietHours(input *QuietHours) *QuietHours {
	if input == nil {
		return nil
	}
	cloned := *input
	return &cloned
}

func cloneTimePtr(input *time.Time) *time.Time {
	if input == nil {
		return nil
	}
	cloned := input.UTC()
	return &cloned
}
