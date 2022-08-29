# Trusted Platform Module 2.0 (TPM)

Trusted Platform Module (TPM, also known as ISO/IEC 11889) is an international standard for a secure cryptoprocessor, a dedicated microcontroller designed to secure hardware through integrated cryptographic keys. The term can also refer to a chip conforming to the standard.


## Add TPM module to virtual machine

Easy way to add TPM to virtual machine is to use Libvirt with Virt-manager

### Create Virtual Machine

After starting virt-manager create new virtual machine

![Create new VM](images/tpm1.png)

### Vedify and edit hardware module list

On the hardware configuration screen, verify list of modules and click ***Add Hardware*** button

![Devices list](images/tpm2.png)

### Add TPM module to VM

From the list of emulated devices choose TPM module and add it to VM

![Add TPM module](images/tpm3.png)

### Finish VM configuration

On the last screen verify once again if TPM module was added properly

![Verify TPM](images/tpm4.png)


## Add TPM emulation to bare metal machine

During applying `#!yaml MachineRegistration` add following key to the yaml `config:elemental:registration:emulate-tpm: true`

```yaml title="registration-tpm.yaml"
--8<-- "examples/quickstart/registration-tpm.yaml"
```

