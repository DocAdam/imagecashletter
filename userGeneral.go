// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package x9

import (
	"fmt"
	"strings"
)

// Errors specific to a UserGeneral Record

// UserGeneral Record
type UserGeneral struct {
	// ID is a client defined string used as a reference to this record.
	ID string `json:"id"`
	// RecordType defines the type of record.
	recordType string
	// OwnerIdentifierIndicator indicates the type of number represented in OwnerIdentifier
	// Values:
	// 0: Not Used
	// 1: Routing Number
	// 2: DUNS Number
	// 3: Federal Tax Identification Number
	// 4: X9 Assignment
	// 5: Other
	OwnerIdentifierIndicator int `json:"ownerIdentifierIndicator"`
	// OwnerIdentifier is a number used by the organization that controls the definition and formatting of this record.
	// Format: Routing Number formats:
	// Applicable when OwnerIdentifierIndicator has Defined Value = 1:
	// TTTTAAAAC where:
	// TTTT: Federal Reserve Prefix
	// AAAA: ABA Institution Identifier
	// C: Check digit
	//
	// DUNS Number format:
	// Applicable when OwnerIdentifierIndicator has Defined Value = 2:
	// XXXXXXXXX where "X" is a numeric value
	//
	// Federal Tax Identification Number format:
	// Applicable when OwnerIdentifierIndicator has Defined Value = 3:
	// XXXXXXXXX where "X" is a numeric value.  The "dash" in the Federal Tax Identification Number
	// (XX-XXXXXXX) is dropped.
	//
	// X9 Assignment
	// Applicable when OwnerIdentifierIndicator has Defined Value = 4: Indicates a Predefined Used Record
	// as defined by X9 within this standard.
	//
	// Other:
	// Applicable when OwnerIdentifierIndicator has Defined Value = 5:
	// Any combination of Alphanumeric special characters agreed to by the exchange partners.
	OwnerIdentifier string `json:"ownerIdentifier"`
	// OwnerIdentifierModifier is a modifier which uniquely identifies the owner within the owning organization.
	OwnerIdentifierModifier string `json:"ownerIdentifierModifier"`
	// UserRecordFormatType uniquely identifies the particular format used to parse and interrogate this record.
	// Provides a means for differentiating user record data layouts. This field shall not be populated with 001
	// since this is reserved for the UserRecordFormatType 001 PayeeEndorsement.
	UserRecordFormatType string `json:"userRecordFormatType"`
	// FormatTypeVersionLevel is a code identifies the version of the UserRecordFormatType. Provides a means for
	// identifying different versions of a record layout.
	FormatTypeVersionLevel string `json:"formatTypeVersionLevel"`
	// LengthUserData is the number of characters or bytes contained in the UserData and must be greater than 0.
	LengthUserData string `json:"LengthUserData"`
	// UserData This field shall be used at the discretion of the owner and exchange partners. The format and structure
	// of this field shall be identified according to OwnerIdentifier, OwnerIdentifierModifier , UserRecordFormatType
	// and FormatTypeVersionLevel.
	UserData string `json:"UserData"`
	// validator is composed for x9 data validation
	validator
	// converters is composed for x9 to golang Converters
	converters
}

// NewUserGeneral returns a new UserGeneral with default values for non exported fields
func NewUserGeneral() *UserGeneral {
	ug := &UserGeneral{
		recordType: "68",
	}
	return ug
}

// Parse takes the input record string and parses the UserGeneral values
func (ug *UserGeneral) Parse(record string) {
	// Character position 1-2, Always "68"
	ug.recordType = "68"
	// 03-03
	ug.OwnerIdentifierIndicator = ug.parseNumField(record[2:3])
	// 04-12
	ug.OwnerIdentifier = ug.parseStringField(record[3:12])
	// 13-32
	ug.OwnerIdentifierModifier = ug.parseStringField(record[12:32])
	// 33-35
	ug.UserRecordFormatType = ug.parseStringField(record[32:35])
	// 36-38
	ug.FormatTypeVersionLevel = ug.parseStringField(record[35:38])
	// 39-45
	ug.LengthUserData = ug.parseStringField(record[38:45])
	// 46-45+(lud)
	ug.UserData = ug.parseStringField(record[45:ug.parseNumField(ug.LengthUserData)])
}

// String writes the UserGeneral struct to a variable length string.
func (ug *UserGeneral) String() string {
	var buf strings.Builder
	buf.Grow(45)
	buf.WriteString(ug.recordType)
	buf.WriteString(ug.OwnerIdentifierIndicatorField())
	buf.WriteString(ug.OwnerIdentifierField())
	buf.WriteString(ug.OwnerIdentifierModifierField())
	buf.WriteString(ug.UserRecordFormatTypeField())
	buf.WriteString(ug.FormatTypeVersionLevelField())
	buf.WriteString(ug.LengthUserDataField())
	buf.Grow(ug.parseNumField(ug.LengthUserData))
	buf.WriteString(ug.UserDataField())
	return buf.String()
}

// Validate performs X9 format rule checks on the record and returns an error if not Validated
// The first error encountered is returned and stops the parsing.
func (ug *UserGeneral) Validate() error {
	if err := ug.fieldInclusion(); err != nil {
		return err
	}
	if ug.recordType != "68" {
		msg := fmt.Sprintf(msgRecordType, 68)
		return &FieldError{FieldName: "recordType", Value: ug.recordType, Msg: msg}
	}
	if ug.UserRecordFormatType == "001" {
		msg := fmt.Sprint(msgInvalid)
		return &FieldError{FieldName: "UserRecordFormatType", Value: ug.UserRecordFormatType, Msg: msg}
	}
	if err := ug.isOwnerIdentifierIndicator(ug.OwnerIdentifierIndicator); err != nil {
		return &FieldError{FieldName: "OwnerIdentifierIndicator",
			Value: ug.OwnerIdentifierIndicatorField(), Msg: err.Error()}
	}
	if ug.OwnerIdentifier != "" {
		if err := ug.isAlphanumericSpecial(ug.OwnerIdentifier); err != nil {
			return &FieldError{FieldName: "OwnerIdentifier", Value: ug.OwnerIdentifier, Msg: err.Error()}
		}
	}
	if ug.OwnerIdentifierModifier != "" {
		if err := ug.isAlphanumericSpecial(ug.OwnerIdentifierModifier); err != nil {
			return &FieldError{FieldName: "OwnerIdentifierModifier",
				Value: ug.OwnerIdentifierModifier, Msg: err.Error()}
		}
	}
	if err := ug.isAlphanumeric(ug.UserRecordFormatType); err != nil {
		return &FieldError{FieldName: "UserRecordFormatType", Value: ug.UserRecordFormatType, Msg: err.Error()}
	}
	if err := ug.isNumeric(ug.FormatTypeVersionLevel); err != nil {
		return &FieldError{FieldName: "FormatTypeVersionLevel",
			Value: ug.FormatTypeVersionLevel, Msg: err.Error()}
	}
	if err := ug.isNumeric(ug.LengthUserData); err != nil {
		return &FieldError{FieldName: "LengthUserData", Value: ug.LengthUserData, Msg: err.Error()}
	}
	if err := ug.isAlphanumericSpecial(ug.UserData); err != nil {
		return &FieldError{FieldName: "UserData", Value: ug.UserData, Msg: err.Error()}
	}
	return nil
}

// fieldInclusion validate mandatory fields are not default values. If fields are
// invalid the Electronic Exchange will be returned.
func (ug *UserGeneral) fieldInclusion() error {
	if ug.recordType == "" {
		return &FieldError{FieldName: "recordType", Value: ug.recordType, Msg: msgFieldInclusion}
	}
	if ug.UserRecordFormatType == "" {
		return &FieldError{FieldName: "UserRecordFormatType",
			Value: ug.UserRecordFormatType, Msg: msgFieldInclusion}
	}
	if ug.FormatTypeVersionLevel == "" {
		return &FieldError{FieldName: "FormatTypeVersionLevel",
			Value: ug.FormatTypeVersionLevel, Msg: msgFieldInclusion}
	}
	if ug.LengthUserData == "" {
		return &FieldError{FieldName: "LengthUserData",
			Value: ug.LengthUserData, Msg: msgFieldInclusion}
	}
	if ug.UserData == "" {
		return &FieldError{FieldName: "UserData",
			Value: ug.UserData, Msg: msgFieldInclusion}
	}
	return nil
}

// OwnerIdentifierIndicatorField gets the OwnerIdentifierIndicator field
func (ug *UserGeneral) OwnerIdentifierIndicatorField() string {
	return ug.numericField(ug.OwnerIdentifierIndicator, 1)
}

// OwnerIdentifierField gets the OwnerIdentifier field
func (ug *UserGeneral) OwnerIdentifierField() string {
	return ug.alphaField(ug.OwnerIdentifier, 9)
}

// OwnerIdentifierModifierField gets the OwnerIdentifierModifier field
func (ug *UserGeneral) OwnerIdentifierModifierField() string {
	return ug.alphaField(ug.OwnerIdentifierModifier, 20)
}

// UserRecordFormatTypeField gets the UserRecordFormatType field
func (ug *UserGeneral) UserRecordFormatTypeField() string {
	return ug.alphaField(ug.UserRecordFormatType, 3)
}

// FormatTypeVersionLevelField gets the FormatTypeVersionLevel field
func (ug *UserGeneral) FormatTypeVersionLevelField() string {
	return ug.alphaField(ug.FormatTypeVersionLevel, 3)
}

// LengthUserDataField gets the LengthUserData field
func (ug *UserGeneral) LengthUserDataField() string {
	return ug.alphaField(ug.LengthUserData, 7)
}

// UserDataField gets the UserData field
func (ug *UserGeneral) UserDataField() string {
	return ug.alphaField(ug.UserData, uint(ug.parseNumField(ug.LengthUserData)))
}
