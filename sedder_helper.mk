VARS = $(shell awk '{ while(match($$0, /\{\{\s*(\w+)\s*\}\}/, var)) { print var[1]; $$0=substr($$0,RSTART+RLENGTH) } }' $(FILE))
PARAMS = $(foreach var,$(VARS), $(sed_template))
check_defined = \
    $(strip $(foreach 1,$1, \
        $(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
    $(if $(value $1),, \
      $(error Undefined variable $1$(if $2, ($2))))
sed_template = -e 's|\{\{\s*$(var)\s*\}\}|$($(var))|g'

# 1-name,2-inputfile,3-prerequisities,4-out
define SEDDER_TARGET_template =
.PHONY: $(1)
$(1): FILE=$(2)
$(1): $(3)
	$$(call check_defined, $$(VARS))
	sed -E $$(PARAMS) $$(FILE) > $(4)
endef
