# google-smart-home-action-go
Build Smart Home Actions for Google Assistant in Go

## Introduction

The Google Smart Home Actions make it possible for the cloud-hosted Google Assistant to control devices built and operated by third parties. The purpose of this library is to simplify the process of building a third party home automation component that the Google Assistant can interact with. This library adds no additional processing capabilities on top of the Google Assistant APIs; it simply wraps the calls to simplify integration into a Go project.

At the moment this library provides no mechanisms to add authentication to your Google Smart Home Action - a separate OAuth capable framework is necessary. Initializing this framework requires a method be supplied to verify tokens - this is code you will write which will look up any Google-supplied tokens against the OAuth framework.

## Google Smart Home Actions

The Google Smart Home framework exists to allow Google Assistant, running on smartphones, smart speakers, etc. to interface with devices made by any number of manufacturers. The Google Assistant is set up in a multitenant architecture; each Google account has its own Assistant profile. This Google account also has its own, distinct object store called a home graph, which contains the profile and state information of every device which all linked Google Smart Home Actions have exposed.

The above is broken down as follows:
- each user has a Google account
- each Google account has a single home graph
- each Google account has zero or more Smart Home Actions linked
- each Smart Home Action linked to a Google account adds zero or more devices to the Google account home graph
- each user, through the Google Assistant, can request a state change of any device on the home graph

The code written in the Smart Home Action is responsible for:
- properly exposing the profile and state of the devices under its control to the home graph
- updating the home graph whenever an underlying device changes
- accepting commands from the home graph to change the state of its devices

# Using this library

This library simplifies the process of handling the Google Smart Home Action code flows. The major entites in the system are:
1. devices. Each device describes a distinct entity that the Google Assistant can interact with. The Google Smart Home Action framework allows for a number of different device types to be specified; depending on the type set the Google Assistant may change how the device is referenced, visualized, etc. The type alone is not sufficient however, it is necessary to define relevant attributes as well (see below).
2. device traits. Each device has a set of traits that describe what the Google Assistant can do with the device. A trait is composed of three distinct components:
  1. attributes, which describe the specific configuration of the trait as it applies to the device
  2. states, which describe the current state of the device
  3. commands, which describe how the Google Assistant is able to control the device with this trait

As an example, a Light is an example of a device which can have 2 traits: OnOff and Brightness.
- The OnOff trait specifies attributes which defines if the device is able to be queried or whether the Google Assistant has to maintain the state internally (some devices are 'write only')
- The OnOff trait specifies the state of on or off
- The OnOff trait specifies the OnOff command which allows the state to be changed.
- The Brightness trait has similiar attributes & states as the OnOff trait, but it specifies a pair of commands which allow the Google Assistant to either define a specific brightness level, or adjust the brightness by a relative amount.

This library has wrappers for some of the most common types of devices and traits to simplify the process of creating a representation of your actual device logic within the home graph. In order to provide flexibility for the consumer of the library, however, it also allows for manual definition of Devices and Traits that can be returned by any API call. Contributions for structured versions of any Smart Home Action device or trait is very much welcome!