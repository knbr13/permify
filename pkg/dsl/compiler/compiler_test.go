package compiler

import (
	"errors"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/Permify/permify/pkg/dsl/parser"
	base "github.com/Permify/permify/pkg/pb/base/v1"
)

// TestCompiler -
func TestCompiler(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "compiler-suite")
}

var _ = Describe("compiler", func() {
	Context("NewCompiler", func() {
		It("Case 1", func() {
			sch, err := parser.NewParser(`
			entity user {}`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(true, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			Expect(is).Should(Equal([]*base.EntityDefinition{
				{
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
			}))
		})

		It("Case 2", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity organization {
				
				relation owner @user
				relation admin @user

				action update = owner or admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(false, sch)

			var is []*base.EntityDefinition
			is, err = c.Compile()
			Expect(err).ShouldNot(HaveOccurred())

			i := []*base.EntityDefinition{
				{
					Name:       "user",
					Relations:  map[string]*base.RelationDefinition{},
					Actions:    map[string]*base.ActionDefinition{},
					References: map[string]base.EntityDefinition_RelationalReference{},
				},
				{
					Name: "organization",
					Actions: map[string]*base.ActionDefinition{
						"update": {
							Name: "update",
							Child: &base.Child{
								Type: &base.Child_Rewrite{
									Rewrite: &base.Rewrite{
										RewriteOperation: base.Rewrite_OPERATION_UNION,
										Children: []*base.Child{
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "owner",
															},
														},
													},
												},
											},
											{
												Type: &base.Child_Leaf{
													Leaf: &base.Leaf{
														Exclusion: false,
														Type: &base.Leaf_ComputedUserSet{
															ComputedUserSet: &base.ComputedUserSet{
																Relation: "admin",
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					Relations: map[string]*base.RelationDefinition{
						"owner": {
							Name: "owner",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
						"admin": {
							Name: "admin",
							RelationReferences: []*base.RelationReference{
								{
									Name: "user",
								},
							},
						},
					},
					References: map[string]base.EntityDefinition_RelationalReference{
						"owner":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"admin":  base.EntityDefinition_RELATIONAL_REFERENCE_RELATION,
						"update": base.EntityDefinition_RELATIONAL_REFERENCE_ACTION,
					},
				},
			}

			Expect(is).Should(Equal(i))
		})

		It("Case 3", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity organization {
				
				relation owner @user
				relation admin @user

				action update = maintainer or admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(false, sch)

			_, err = c.Compile()
			Expect(err).Should(Equal(errors.New(base.ErrorCode_ERROR_CODE_UNDEFINED_RELATION_REFERENCE.String())))
		})

		It("Case 4", func() {
			sch, err := parser.NewParser(`
			entity user {}
				
			entity parent {
				
				relation admin @user
			}

			entity organization {
				
				relation parent @parent
				relation admin @user
			}

			entity repository {
				
				relation parent @organization
				action update = parent.parent.admin or admin
			}
			`).Parse()

			Expect(err).ShouldNot(HaveOccurred())

			c := NewCompiler(false, sch)

			_, err = c.Compile()
			Expect(err).Should(Equal(errors.New(base.ErrorCode_ERROR_CODE_NOT_SUPPORTED_RELATION_WALK.String())))
		})
	})
})
